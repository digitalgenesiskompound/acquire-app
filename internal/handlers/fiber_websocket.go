package handlers

import (
	"fmt"
	"net/http"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

// HandleFiberWebSocket provides WebSocket endpoint information for Fiber
func (h *WebusbHandler) HandleFiberWebSocket(c *fiber.Ctx) error {
	acquisitionID := c.Params("acquisitionId")
	if acquisitionID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing acquisition ID",
		})
	}

	// Validate acquisition exists
	_, err := h.sessionManager.GetAcquisition(acquisitionID)
	if err != nil {
		slog.Error("Acquisition not found", "acquisitionId", acquisitionID, "error", err)
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Acquisition not found",
		})
	}

	// For now, return WebSocket connection information
	// In production, you would either:
	// 1. Use a WebSocket library compatible with Fiber
	// 2. Set up a separate WebSocket server on a different port
	// 3. Use Fiber's WebSocket middleware (if available)
	
	baseURL := fmt.Sprintf("%s://%s", c.Protocol(), c.Hostname())
	if port := c.Port(); port != "" {
		baseURL += ":" + port
	}
	
	return c.JSON(fiber.Map{
		"message": "WebSocket endpoint available",
		"acquisitionId": acquisitionID,
		"wsUrl": fmt.Sprintf("ws://%s/api/webusb/stream/%s", c.Hostname(), acquisitionID),
		"protocol": "WebSocket required for real-time data streaming",
		"instructions": "Use a WebSocket client to connect to this endpoint for real-time data transfer",
	})
}

// CreateWebSocketRoute creates a WebSocket-compatible route that can be used with a separate HTTP server
func CreateWebSocketRoute(webusbHandler *WebusbHandler) http.HandlerFunc {
	wsHandler := NewWebSocketHandler(webusbHandler.GetSessionManager())
	return wsHandler.HandleWebSocket
}
