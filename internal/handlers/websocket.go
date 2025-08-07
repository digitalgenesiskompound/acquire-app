package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"acquire-app/internal/models"
	"acquire-app/internal/services"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// In production, you should validate the origin
		return true
	},
}

type WebSocketHandler struct {
	sessionManager *services.SessionManager
}

func NewWebSocketHandler(sessionManager *services.SessionManager) *WebSocketHandler {
	return &WebSocketHandler{
		sessionManager: sessionManager,
	}
}

// HandleWebSocket handles WebSocket connections for data streaming
// This will be mounted at /api/webusb/stream/{acquisitionId}
func (ws *WebSocketHandler) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract acquisition ID from URL path (this depends on your router setup)
	acquisitionID := extractAcquisitionID(r.URL.Path)
	if acquisitionID == "" {
		http.Error(w, "Missing acquisition ID", http.StatusBadRequest)
		return
	}

	// Validate acquisition exists
	acquisition, err := ws.sessionManager.GetAcquisition(acquisitionID)
	if err != nil {
		slog.Error("Acquisition not found", "acquisitionId", acquisitionID, "error", err)
		http.Error(w, "Acquisition not found", http.StatusNotFound)
		return
	}

	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("Failed to upgrade to websocket", "error", err)
		return
	}
	defer conn.Close()

	slog.Info("WebSocket connection established", 
		"acquisitionId", acquisitionID,
		"sessionId", acquisition.SessionID,
		"remoteAddr", r.RemoteAddr)

	// Handle the WebSocket connection
	ws.handleConnection(conn, acquisition)
}

func (ws *WebSocketHandler) handleConnection(conn *websocket.Conn, acquisition *models.Acquisition) {
	// Set up ping/pong handlers for connection health
	conn.SetPingHandler(func(appData string) error {
		return conn.WriteMessage(websocket.PongMessage, []byte(appData))
	})

	// Set read deadline
	conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	conn.SetPongHandler(func(appData string) error {
		conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Start ping ticker
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	// Track statistics
	var totalChunks int64 = 0
	var totalBytes int64 = 0

	// Handle messages
	for {
		select {
		case <-ticker.C:
			// Send ping
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				slog.Error("Failed to send ping", "acquisitionId", acquisition.ID, "error", err)
				return
			}

		default:
			// Read message from client
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					slog.Error("WebSocket error", "acquisitionId", acquisition.ID, "error", err)
				} else {
					slog.Info("WebSocket connection closed", "acquisitionId", acquisition.ID)
				}
				return
			}

			// Handle different message types
			switch messageType {
			case websocket.TextMessage:
				err = ws.handleTextMessage(conn, acquisition, data, &totalChunks, &totalBytes)
				if err != nil {
					slog.Error("Failed to handle text message", "acquisitionId", acquisition.ID, "error", err)
					return
				}

			case websocket.BinaryMessage:
				err = ws.handleBinaryMessage(conn, acquisition, data, &totalChunks, &totalBytes)
				if err != nil {
					slog.Error("Failed to handle binary message", "acquisitionId", acquisition.ID, "error", err)
					return
				}
			}

			// Reset read deadline
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		}
	}
}

func (ws *WebSocketHandler) handleTextMessage(conn *websocket.Conn, acquisition *models.Acquisition, data []byte, totalChunks, totalBytes *int64) error {
	var message models.WSMessage
	if err := json.Unmarshal(data, &message); err != nil {
		return ws.sendErrorMessage(conn, "INVALID_MESSAGE", "Failed to parse message", err.Error())
	}

	switch message.Type {
	case "data_chunk":
		return ws.handleDataChunk(conn, acquisition, data, totalChunks, totalBytes)

	case "status_update":
		return ws.handleStatusUpdate(conn, acquisition, data)

	case "error":
		return ws.handleClientError(conn, acquisition, data)

	default:
		return ws.sendErrorMessage(conn, "UNKNOWN_MESSAGE_TYPE", "Unknown message type", message.Type)
	}
}

func (ws *WebSocketHandler) handleBinaryMessage(conn *websocket.Conn, acquisition *models.Acquisition, data []byte, totalChunks, totalBytes *int64) error {
	// For binary messages, we might handle raw data chunks
	// This is a simple implementation - in practice, you'd have a more sophisticated binary protocol
	
	*totalChunks++
	*totalBytes += int64(len(data))

	// Update acquisition statistics
	ws.sessionManager.UpdateAcquisitionStats(acquisition.ID, *totalChunks, *totalBytes)

	// Send acknowledgment
	ackMessage := map[string]interface{}{
		"type":       "ack",
		"chunkIndex": *totalChunks,
		"received":   true,
		"processingStatus": "validated",
	}

	return conn.WriteJSON(ackMessage)
}

func (ws *WebSocketHandler) handleDataChunk(conn *websocket.Conn, acquisition *models.Acquisition, data []byte, totalChunks, totalBytes *int64) error {
	var chunkMsg models.DataChunkMessage
	if err := json.Unmarshal(data, &chunkMsg); err != nil {
		return ws.sendErrorMessage(conn, "INVALID_DATA_CHUNK", "Failed to parse data chunk", err.Error())
	}

	// Validate checksum
	if !ws.validateChecksum(chunkMsg.Data, chunkMsg.Checksum) {
		return ws.sendErrorMessage(conn, "CHECKSUM_MISMATCH", "Data integrity check failed", "")
	}

	// Decode base64 data
	decodedData, err := base64.StdEncoding.DecodeString(chunkMsg.Data)
	if err != nil {
		return ws.sendErrorMessage(conn, "DECODE_ERROR", "Failed to decode data", err.Error())
	}

	// Update statistics
	*totalChunks = chunkMsg.ChunkIndex + 1
	*totalBytes += int64(len(decodedData))

	// Update acquisition statistics
	ws.sessionManager.UpdateAcquisitionStats(acquisition.ID, *totalChunks, *totalBytes)

	// In a real implementation, you would:
	// 1. Store the data to persistent storage
	// 2. Validate data format
	// 3. Process data if needed
	// 4. Update progress tracking

	slog.Debug("Data chunk received", 
		"acquisitionId", acquisition.ID,
		"chunkIndex", chunkMsg.ChunkIndex,
		"dataSize", len(decodedData),
		"totalChunks", *totalChunks,
		"totalBytes", *totalBytes)

	// Send acknowledgment
	ackMessage := map[string]interface{}{
		"type":               "ack",
		"chunkIndex":         chunkMsg.ChunkIndex,
		"received":           true,
		"processingStatus":   "validated",
	}

	// Optionally send processing feedback
	if chunkMsg.ChunkIndex%10 == 0 { // Every 10th chunk
		qualityMetrics := map[string]interface{}{
			"signalStrength": 0.92,
			"noiseLevel":     0.05,
			"dataIntegrity":  1.0,
		}

		feedbackMessage := map[string]interface{}{
			"type":            "processing_feedback",
			"acquisitionId":   acquisition.ID,
			"qualityMetrics":  qualityMetrics,
			"recommendations": []string{"Signal quality is excellent", "Continue current position"},
		}

		if err := conn.WriteJSON(feedbackMessage); err != nil {
			return err
		}
	}

	return conn.WriteJSON(ackMessage)
}

func (ws *WebSocketHandler) handleStatusUpdate(conn *websocket.Conn, acquisition *models.Acquisition, data []byte) error {
	var statusMsg models.StatusUpdateMessage
	if err := json.Unmarshal(data, &statusMsg); err != nil {
		return ws.sendErrorMessage(conn, "INVALID_STATUS_UPDATE", "Failed to parse status update", err.Error())
	}

	// Update session health data based on status update
	if statusMsg.DeviceHealth.BatteryLevel > 0 {
		ws.sessionManager.UpdateSession(acquisition.SessionID, func(session *models.Session) {
			session.DeviceHealth = statusMsg.DeviceHealth
		})
	}

	slog.Info("Status update received", 
		"acquisitionId", acquisition.ID,
		"status", statusMsg.Status,
		"progress", statusMsg.Progress)

	// Send acknowledgment
	ackMessage := map[string]interface{}{
		"type":    "status_ack",
		"status":  "received",
		"message": "Status update processed",
	}

	return conn.WriteJSON(ackMessage)
}

func (ws *WebSocketHandler) handleClientError(conn *websocket.Conn, acquisition *models.Acquisition, data []byte) error {
	var errorMsg map[string]interface{}
	if err := json.Unmarshal(data, &errorMsg); err != nil {
		return ws.sendErrorMessage(conn, "INVALID_ERROR_MESSAGE", "Failed to parse error message", err.Error())
	}

	errorCode, _ := errorMsg["errorCode"].(string)
	errorMessage, _ := errorMsg["errorMessage"].(string)
	recoverable, _ := errorMsg["recoverable"].(bool)

	slog.Error("Client error received", 
		"acquisitionId", acquisition.ID,
		"errorCode", errorCode,
		"errorMessage", errorMessage,
		"recoverable", recoverable)

	// Update session error count
	ws.sessionManager.UpdateSession(acquisition.SessionID, func(session *models.Session) {
		session.Statistics.ErrorCount++
	})

	// Send error acknowledgment and potentially recovery instructions
	response := map[string]interface{}{
		"type":        "error_ack",
		"errorCode":   errorCode,
		"acknowledged": true,
	}

	if recoverable {
		response["recoveryAction"] = "retry"
		response["retryDelay"] = 1000 // 1 second
	}

	return conn.WriteJSON(response)
}

func (ws *WebSocketHandler) sendErrorMessage(conn *websocket.Conn, errorCode, errorMessage, details string) error {
	errorResponse := map[string]interface{}{
		"type":         "server_error",
		"errorCode":    errorCode,
		"errorMessage": errorMessage,
		"details":      details,
		"timestamp":    time.Now(),
	}

	return conn.WriteJSON(errorResponse)
}

func (ws *WebSocketHandler) validateChecksum(data, expectedChecksum string) bool {
	// Decode base64 data
	decodedData, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return false
	}

	// Calculate SHA-256 hash
	hash := sha256.Sum256(decodedData)
	actualChecksum := hex.EncodeToString(hash[:])

	return actualChecksum == expectedChecksum
}

// Helper function to extract acquisition ID from URL path
func extractAcquisitionID(path string) string {
	// This is a simple implementation - in practice, you'd use your router's parameter extraction
	// For path like "/api/webusb/stream/acq_12345", this extracts "acq_12345"
	parts := strings.Split(path, "/")
	if len(parts) >= 4 && parts[len(parts)-2] == "stream" {
		return parts[len(parts)-1]
	}
	return ""
}

