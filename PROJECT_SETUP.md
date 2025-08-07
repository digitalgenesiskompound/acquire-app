# Acquire-App Project Setup

## Overview
This is a Go-based web application project called "acquire-app" that has been initialized with Go modules and set up with modern web framework dependencies.

## Go Module Configuration

### Module Information
- **Module Name**: `acquire-app`
- **Go Version**: 1.24 (latest available: 1.24.5)
- **Module File**: `go.mod`

### Dependencies

#### Direct Dependencies
- **Fiber v2** (`github.com/gofiber/fiber/v2 v2.52.9`) - Modern Express-inspired web framework for Go
- **Gorilla Mux** (`github.com/gorilla/mux v1.8.1`) - HTTP router and URL matcher

#### Indirect Dependencies
The following packages are automatically managed as transitive dependencies:
- `github.com/andybalholm/brotli v1.1.0` - Brotli compression
- `github.com/google/uuid v1.6.0` - UUID generation
- `github.com/klauspost/compress v1.17.9` - Compression algorithms
- `github.com/mattn/go-colorable v0.1.13` - Cross-platform colored output
- `github.com/mattn/go-isatty v0.0.20` - Terminal detection
- `github.com/mattn/go-runewidth v0.0.16` - Unicode width calculation
- `github.com/rivo/uniseg v0.2.0` - Unicode segmentation
- `github.com/valyala/bytebufferpool v1.0.0` - Byte buffer pool
- `github.com/valyala/fasthttp v1.51.0` - Fast HTTP implementation
- `github.com/valyala/tcplisten v1.0.0` - TCP listener utilities
- `golang.org/x/sys v0.28.0` - System call interface

## Development Environment
- **Go Version Installed**: 1.24.5 (via snap)
- **Platform**: Linux/AMD64
- **Package Manager**: Go Modules
- **Dependency Files**: `go.mod` and `go.sum`

## Next Steps
The project is now ready for development with:
1. Go modules properly initialized
2. Fiber v2 web framework available for building HTTP APIs
3. All dependencies downloaded and verified
4. Ready for creating web server endpoints and middleware

## Step 3: Fiber Server Implementation ✅

**COMPLETED**: Implemented `cmd/server/main.go` with the following features:

### Server Implementation
- **Framework**: Uses Fiber v2 (replaced Gorilla Mux)
- **Static Files**: Serves all files under `/web` directory at root path `/`
- **Binding**: Listens on `0.0.0.0:8080` by default
- **Environment Variables**: 
  - `HOST` - Server host (default: "0.0.0.0")
  - `PORT` - Server port (default: "8080")
  - `ENV` - Environment (default: "development")
  - `DEBUG` - Debug mode (default: true)

### Features Implemented
1. **Structured Logging**: Uses Go's `log/slog` package with JSON output
2. **Graceful Shutdown**: Handles SIGTERM/SIGINT with 30-second timeout
3. **Middleware**: 
   - Recovery middleware for panic handling
   - Request logging middleware
   - Error handling with structured logging
4. **Health Check**: `/health` endpoint returns JSON status and timestamp
5. **Static File Serving**: All files under `/web` served at root path
6. **WebUSB Placeholder**: TODO comments for future WebUSB API endpoints:
   - `POST /api/webusb/connect`
   - `POST /api/webusb/disconnect` 
   - `POST /api/webusb/transfer`
   - `GET /api/webusb/devices`

### Files Modified/Created
- `cmd/server/main.go` - Complete rewrite using Fiber v2
- `internal/config/config.go` - Added Host field support
- `web/index.html` - Created main HTML file with WebUSB placeholder

### Testing Verified
- Server builds successfully with `go build cmd/server/main.go`
- Health check endpoint responds correctly at `/health`
- Environment variables work (tested with HOST=127.0.0.1 PORT=9000)
- Graceful shutdown works with SIGTERM/SIGINT signals
- Static files served from `/web` directory

## Notes
- The project maintains both Fiber v2 and Gorilla Mux for flexibility
- All dependencies are pinned to specific versions for reproducible builds
- The `go.sum` file ensures dependency integrity and security
