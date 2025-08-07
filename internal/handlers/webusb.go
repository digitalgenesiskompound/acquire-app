package handlers

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"acquire-app/internal/models"
	"acquire-app/internal/services"
)

type WebusbHandler struct {
	sessionManager *services.SessionManager
}

func NewWebusbHandler() *WebusbHandler {
	return &WebusbHandler{
		sessionManager: services.NewSessionManager(),
	}
}

// RegisterDevice handles POST /api/webusb/devices/register
func (h *WebusbHandler) RegisterDevice(c *fiber.Ctx) error {
	var req models.DeviceRegistrationRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse device registration request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
	}

	// Validate required fields
	if req.DeviceInfo.ProductName == "" || req.DeviceInfo.SerialNumber == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Missing required device information",
			Code:    "MISSING_DEVICE_INFO",
			Details: "productName and serialNumber are required",
		})
	}

	// Create new session
	session, err := h.sessionManager.CreateSession(req.DeviceInfo, req.Capabilities)
	if err != nil {
		slog.Error("Failed to create session", "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Failed to create session",
			Code:    "SESSION_CREATE_ERROR",
			Details: err.Error(),
		})
	}

	// Prepare response
	response := models.DeviceRegistrationResponse{
		Success:   true,
		SessionID: session.ID,
		DeviceID:  session.DeviceID,
		ServerConfig: models.ServerConfig{
			BufferSize:         8192,
			Timeout:            5000,
			CompressionEnabled: true,
		},
		AcquisitionSettings: models.AcquisitionSettings{
			SampleRate: 44100,
			BitDepth:   16,
			Channels:   1,
		},
	}

	slog.Info("Device registered successfully", 
		"sessionId", session.ID, 
		"deviceId", session.DeviceID,
		"productName", req.DeviceInfo.ProductName)

	return c.JSON(response)
}

// ConnectDevice handles POST /api/webusb/devices/connect
func (h *WebusbHandler) ConnectDevice(c *fiber.Ctx) error {
	var req models.DeviceConnectionRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse device connection request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
	}

	// Validate session exists
	session, err := h.sessionManager.GetSession(req.SessionID)
	if err != nil {
		slog.Error("Session not found", "sessionId", req.SessionID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Session not found",
			Code:    "SESSION_NOT_FOUND",
			Details: err.Error(),
		})
	}

	// Update session with connection status
	err = h.sessionManager.UpdateSession(req.SessionID, func(s *models.Session) {
		s.Status = "active"
		s.DeviceConnected = req.ConnectionStatus.Connected
		s.DeviceHealth.Temperature = req.DeviceState.Temperature
		s.DeviceHealth.BatteryLevel = req.DeviceState.BatteryLevel
		s.DeviceHealth.LastHealthCheck = time.Now()
	})

	if err != nil {
		slog.Error("Failed to update session", "sessionId", req.SessionID, "error", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Error:   "Failed to update session",
			Code:    "SESSION_UPDATE_ERROR",
			Details: err.Error(),
		})
	}

	// Determine next action based on device state
	nextAction := "ready"
	calibrationRequired := false
	estimatedCalibrationTime := 0

	if !req.DeviceState.Calibrated && session.Capabilities.HasCalibration {
		nextAction = "calibration"
		calibrationRequired = true
		estimatedCalibrationTime = 30
	}

	response := models.DeviceConnectionResponse{
		Success:                  true,
		Message:                  "Device connection confirmed",
		NextAction:               nextAction,
		CalibrationRequired:      calibrationRequired,
		EstimatedCalibrationTime: estimatedCalibrationTime,
	}

	slog.Info("Device connected successfully", 
		"sessionId", req.SessionID, 
		"deviceId", req.DeviceID,
		"nextAction", nextAction)

	return c.JSON(response)
}

// DisconnectDevice handles POST /api/webusb/devices/disconnect
func (h *WebusbHandler) DisconnectDevice(c *fiber.Ctx) error {
	var req models.DeviceDisconnectionRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse device disconnection request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
	}

	// Close session
	err := h.sessionManager.CloseSession(req.SessionID)
	if err != nil {
		slog.Error("Failed to close session", "sessionId", req.SessionID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Session not found",
			Code:    "SESSION_NOT_FOUND",
			Details: err.Error(),
		})
	}

	response := models.DeviceDisconnectionResponse{
		Success:       true,
		Message:       "Device disconnected successfully",
		SessionClosed: true,
		DataPreserved: true,
	}

	slog.Info("Device disconnected successfully", 
		"sessionId", req.SessionID, 
		"deviceId", req.DeviceID,
		"reason", req.DisconnectionReason)

	return c.JSON(response)
}

// StartAcquisition handles POST /api/webusb/acquisition/start
func (h *WebusbHandler) StartAcquisition(c *fiber.Ctx) error {
	var req models.AcquisitionStartRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse acquisition start request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
	}

	// Validate session exists
	_, err := h.sessionManager.GetSession(req.SessionID)
	if err != nil {
		slog.Error("Session not found", "sessionId", req.SessionID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Session not found",
			Code:    "SESSION_NOT_FOUND",
			Details: err.Error(),
		})
	}

	// Create acquisition
	acquisition, err := h.sessionManager.CreateAcquisition(req.SessionID, req.AcquisitionParams, req.Metadata)
	if err != nil {
		slog.Error("Failed to create acquisition", "sessionId", req.SessionID, "error", err)
		return c.Status(fiber.StatusConflict).JSON(models.ErrorResponse{
			Error:   "Failed to create acquisition",
			Code:    "ACQUISITION_CREATE_ERROR",
			Details: err.Error(),
		})
	}

	// Build WebSocket endpoint URL
	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())
	if port := c.Port(); port != "" {
		baseURL += ":" + port
	}
	streamEndpoint := fmt.Sprintf("ws://%s/api/webusb/stream/%s", c.Hostname(), acquisition.ID)

	response := models.AcquisitionStartResponse{
		Success:          true,
		AcquisitionID:    acquisition.ID,
		StreamEndpoint:   streamEndpoint,
		ExpectedDataSize: 10485760, // 10MB default
		ChunkSize:        4096,     // 4KB chunks
	}

	slog.Info("Acquisition started successfully", 
		"acquisitionId", acquisition.ID, 
		"sessionId", req.SessionID,
		"mode", req.AcquisitionParams.Mode)

	return c.JSON(response)
}

// StopAcquisition handles POST /api/webusb/acquisition/stop
func (h *WebusbHandler) StopAcquisition(c *fiber.Ctx) error {
	var req models.AcquisitionStopRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse acquisition stop request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
	}

	// Stop acquisition
	acquisition, err := h.sessionManager.StopAcquisition(req.AcquisitionID, req.Reason)
	if err != nil {
		slog.Error("Failed to stop acquisition", "acquisitionId", req.AcquisitionID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Failed to stop acquisition",
			Code:    "ACQUISITION_STOP_ERROR",
			Details: err.Error(),
		})
	}

	response := models.AcquisitionStopResponse{
		Success:      true,
		Message:      "Acquisition stopped successfully",
		FinalStats:   acquisition.Statistics,
		DataLocation: fmt.Sprintf("/api/webusb/acquisition/%s/data", req.AcquisitionID),
	}

	slog.Info("Acquisition stopped successfully", 
		"acquisitionId", req.AcquisitionID, 
		"reason", req.Reason,
		"totalBytes", acquisition.Statistics.TotalBytes)

	return c.JSON(response)
}

// GetSessionStatus handles GET /api/webusb/sessions/{sessionId}/status
func (h *WebusbHandler) GetSessionStatus(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Missing session ID",
			Code:    "MISSING_SESSION_ID",
			Details: "Session ID is required in the URL path",
		})
	}

	session, err := h.sessionManager.GetSession(sessionID)
	if err != nil {
		slog.Error("Session not found", "sessionId", sessionID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Session not found",
			Code:    "SESSION_NOT_FOUND",
			Details: err.Error(),
		})
	}

	response := models.SessionStatusResponse{
		SessionID:          session.ID,
		Status:             session.Status,
		DeviceConnected:    session.DeviceConnected,
		CurrentAcquisition: session.CurrentAcquisition,
		StartTime:          session.StartTime,
		LastActivity:       session.LastActivity,
		Statistics:         session.Statistics,
		DeviceHealth:       session.DeviceHealth,
	}

	return c.JSON(response)
}

// ProcessHeartbeat handles POST /api/webusb/sessions/{sessionId}/heartbeat
func (h *WebusbHandler) ProcessHeartbeat(c *fiber.Ctx) error {
	sessionID := c.Params("sessionId")
	if sessionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Missing session ID",
			Code:    "MISSING_SESSION_ID",
			Details: "Session ID is required in the URL path",
		})
	}

	var req models.HeartbeatRequest
	if err := c.BodyParser(&req); err != nil {
		slog.Error("Failed to parse heartbeat request", "error", err)
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Error:   "Invalid request body",
			Code:    "INVALID_REQUEST",
			Details: err.Error(),
		})
	}

	// Process heartbeat
	err := h.sessionManager.ProcessHeartbeat(sessionID, req.ClientState)
	if err != nil {
		slog.Error("Failed to process heartbeat", "sessionId", sessionID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(models.ErrorResponse{
			Error:   "Session not found",
			Code:    "SESSION_NOT_FOUND",
			Details: err.Error(),
		})
	}

	// Prepare server state and instructions
	serverState := models.ServerState{
		ProcessingQueue:    len(h.sessionManager.GetActiveAcquisitions()),
		StorageUtilization: 0.23, // This would be calculated from actual storage usage
		SystemHealth:       "optimal",
	}

	instructions := []models.Instruction{}
	
	// Add buffer size adjustment instruction if buffer utilization is high
	if req.ClientState.BufferUtilization > 0.8 {
		instructions = append(instructions, models.Instruction{
			Action:  "adjust_buffer_size",
			NewSize: 16384,
		})
	}

	response := models.HeartbeatResponse{
		Success:      true,
		ServerState:  serverState,
		Instructions: instructions,
	}

	return c.JSON(response)
}

// Helper methods for WebSocket support (to be used with gorilla/websocket)
func (h *WebusbHandler) GetSessionManager() *services.SessionManager {
	return h.sessionManager
}

// Cleanup expired sessions - can be called periodically
func (h *WebusbHandler) CleanupExpiredSessions() {
	timeout := 1 * time.Hour // 1 hour timeout
	cleaned := h.sessionManager.CleanupExpiredSessions(timeout)
	if cleaned > 0 {
		slog.Info("Cleaned up expired sessions", "count", cleaned)
	}
}
