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
	"acquire-app/internal/handlers"
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

	// Initialize WebUSB handler
	webusbHandler := handlers.NewWebusbHandler()

	// WebUSB API endpoints for device management
	api := app.Group("/api/webusb")
	
	// Device management endpoints
	api.Post("/devices/register", webusbHandler.RegisterDevice)
	api.Post("/devices/connect", webusbHandler.ConnectDevice)
	api.Post("/devices/disconnect", webusbHandler.DisconnectDevice)
	
	// Data acquisition endpoints
	api.Post("/acquisition/start", webusbHandler.StartAcquisition)
	api.Post("/acquisition/stop", webusbHandler.StopAcquisition)
	
	// Session management endpoints
	api.Get("/sessions/:sessionId/status", webusbHandler.GetSessionStatus)
	api.Post("/sessions/:sessionId/heartbeat", webusbHandler.ProcessHeartbeat)
	
	// WebSocket endpoint for real-time data streaming
	app.Get("/api/webusb/stream/:acquisitionId", webusbHandler.HandleFiberWebSocket)
	
	// Start session cleanup goroutine
	go func() {
		ticker := time.NewTicker(15 * time.Minute) // Cleanup every 15 minutes
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				webusbHandler.CleanupExpiredSessions()
			}
		}
	}()

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

	// Server addresses
	httpsAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)
	httpAddr := fmt.Sprintf("%s:%s", cfg.Host, cfg.HTTPPort)
	
	// Check for SSL certificates
	certFile := "./certs/server.crt"
	keyFile := "./certs/server.key"
	
	// Check if we're in Docker (certs at /certs) or local dev (certs at ./certs)
	if _, err := os.Stat("/certs/server.crt"); err == nil {
		certFile = "/certs/server.crt"
		keyFile = "/certs/server.key"
		slog.Info("Using Docker certificate directory", "cert", certFile, "key", keyFile)
	}
	
	useHTTPS := false
	if _, err := os.Stat(certFile); err == nil {
		if _, err := os.Stat(keyFile); err == nil {
			useHTTPS = true
			slog.Info("SSL certificates found, enabling HTTPS", "cert", certFile, "key", keyFile)
		}
	}
	
	// Create a second Fiber app for HTTP (if different from HTTPS port)
	var httpApp *fiber.App
	runBothServers := useHTTPS && cfg.Port != cfg.HTTPPort
	
	if runBothServers {
		// Clone the main app configuration for HTTP server
		httpApp = fiber.New(fiber.Config{
			ServerHeader: "Acquire-App-HTTP",
			ErrorHandler: func(c *fiber.Ctx, err error) error {
				slog.Error("HTTP Request error", "error", err.Error(), "path", c.Path())
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal Server Error",
				})
			},
		})
		
		// Add same middleware to HTTP app
		httpApp.Use(recover.New())
		httpApp.Use(fiberlogger.New(fiberlogger.Config{
			Format: "HTTP ${time} | ${status} | ${latency} | ${ip} | ${method} | ${path} | ${error}\n",
		}))
		
		// Add health check to HTTP app
		httpApp.Get("/health", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{
				"status": "ok",
				"protocol": "http",
				"timestamp": time.Now().Unix(),
			})
		})
		
		// Add redirect to HTTPS for WebUSB endpoints
		httpApp.All("/api/webusb/*", func(c *fiber.Ctx) error {
			httpsURL := fmt.Sprintf("https://%s:%s%s", cfg.Host, cfg.Port, c.OriginalURL())
			return c.Redirect(httpsURL, 301)
		})
		
		// Serve static files on HTTP as well
		httpApp.Static("/", webDir)
	}
	
	if useHTTPS {
		slog.Info("Starting HTTPS server", 
			"host", cfg.Host, 
			"port", cfg.Port, 
			"environment", cfg.Environment,
			"address", "https://"+httpsAddr,
			"cert", certFile,
			"key", keyFile)
			
		if runBothServers {
			slog.Info("Starting HTTP redirect server", 
				"host", cfg.Host, 
				"port", cfg.HTTPPort, 
				"address", "http://"+httpAddr,
				"redirects_to", "https://"+httpsAddr)
		}
	} else {
		slog.Info("Starting HTTP server (no SSL certificates found)", 
			"host", cfg.Host, 
			"port", cfg.Port, 
			"environment", cfg.Environment,
			"address", "http://"+httpsAddr)
	}

	// Start servers in goroutines
	if useHTTPS {
		// Start HTTPS server
		go func() {
			if err := app.ListenTLS(httpsAddr, certFile, keyFile); err != nil {
				slog.Error("Failed to start HTTPS server", "error", err)
				os.Exit(1)
			}
		}()
		
		// Start HTTP redirect server if running both
		if runBothServers {
			go func() {
				if err := httpApp.Listen(httpAddr); err != nil {
					slog.Error("Failed to start HTTP redirect server", "error", err)
					os.Exit(1)
				}
			}()
		}
	} else {
		// Start HTTP server only
		go func() {
			if err := app.Listen(httpsAddr); err != nil {
				slog.Error("Failed to start HTTP server", "error", err)
				os.Exit(1)
			}
		}()
	}

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
