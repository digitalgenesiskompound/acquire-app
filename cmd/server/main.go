package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"acquire-app/internal/config"
)

func main() {
	// Initialize structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load configuration
	cfg := config.Load()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		ServerHeader: "Acquire-App",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			slog.Error("Request error", "error", err.Error(), "path", c.Path())
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})

	// Add middleware
	app.Use(recover.New())
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format: "${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
	}))

	// Health check endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// TODO: Add WebUSB API endpoints here
	// Future endpoints for WebUSB communication:
	// - POST /api/webusb/connect
	// - POST /api/webusb/disconnect
	// - POST /api/webusb/transfer
	// - GET /api/webusb/devices

	// Serve all files under /web directory
	// Check if we're in Docker (web files at /web) or local dev (web files at ./web)
	webDir := "./web"
	if _, err := os.Stat("/web"); err == nil {
		webDir = "/web"
		slog.Info("Using Docker web directory", "path", webDir)
	} else {
		slog.Info("Using local web directory", "path", webDir)
	}
	app.Static("/", webDir)

	// Server address
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	slog.Info("Starting server", 
		"host", cfg.Host, 
		"port", cfg.Port, 
		"environment", cfg.Environment,
		"address", addr)

	// Start server in a goroutine
	go func() {
		if err := app.Listen(addr); err != nil {
			slog.Error("Failed to start server", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Block until we receive a signal
	<-c

	slog.Info("Shutting down server...")

	// Create a context with a timeout for graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Attempt graceful shutdown
	if err := app.ShutdownWithContext(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
		os.Exit(1)
	}

	slog.Info("Server exited gracefully")
}
