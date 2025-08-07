# Acquire-App

A comprehensive Go web application designed for intra-oral data capture and processing, featuring modern web frameworks, Docker containerization, and WebUSB integration capabilities.

## 📁 Project Structure

The project follows Go best practices with a clean, modular directory structure:

```
Acquire-App/
├── cmd/
│   └── server/
│       └── main.go          # Main server entry point with Fiber v2
├── internal/                # Private application code (Go convention)
│   ├── config/
│   │   └── config.go        # Environment-based configuration management
│   └── handlers/
│       └── handlers.go      # HTTP request handlers (legacy, not currently used)
├── web/                     # Static web assets served at root path
│   ├── css/
│   │   └── style.css        # Discord-dark theme styling
│   ├── js/
│   │   └── app.js          # Client-side JavaScript with WebUSB placeholders
│   └── index.html          # Main HTML interface for intra-oral capture
├── bin/
│   └── server              # Compiled binary (created after build)
├── .dockerignore           # Files excluded from Docker build context
├── Dockerfile              # Multi-stage Docker build configuration
├── docker-compose.yml      # Container orchestration setup
├── go.mod                  # Go module definition (acquire-app, Go 1.24)
├── go.sum                  # Dependency checksums for security
├── PROJECT_OVERVIEW.md     # Detailed project documentation
├── PROJECT_SETUP.md        # Development setup and implementation notes
└── README.md              # This comprehensive guide
```

## 🏗️ Architecture & Technology Stack

### Backend (Go)
- **Framework**: [Fiber v2](https://github.com/gofiber/fiber) - Modern Express-inspired web framework
- **Routing**: [Gorilla Mux](https://github.com/gorilla/mux) v1.8.1 (available for complex routing needs)
- **Configuration**: Environment variable-based with sensible defaults
- **Logging**: Structured JSON logging using Go's `log/slog` package
- **Graceful Shutdown**: 30-second timeout with signal handling
- **Go Version**: 1.24 (latest available)

### Frontend
- **HTML5**: Semantic, accessible markup
- **CSS3**: Discord-dark theme with modern flexbox layouts
- **JavaScript**: ES6+ with WebUSB API integration (planned)
- **Static Serving**: All web assets served from `/web` directory

### Infrastructure
- **Containerization**: Multi-stage Docker builds for optimized images
- **Orchestration**: Docker Compose for development and deployment
- **Base Images**: Alpine Linux (builder) + Distroless (runtime)
- **Security**: Non-root container execution

## 🚀 Getting Started

### Prerequisites

- **Go**: Version 1.24 or later
- **Docker**: Latest version with Docker Compose
- **Git**: For version control

### Quick Start with Docker Compose (Recommended)

This is the fastest way to get the application running:

```bash
# Clone the repository
git clone <repository-url>
cd Acquire-App

# Build and start the application
docker compose up --build
```

**Important**: Use `docker compose up --build` to ensure container takes on file changes from the directory.

The application will be available at: **http://localhost:8080**

### Local Development Setup

For active development with hot reloading:

```bash
# Install dependencies
go mod tidy

# Run the development server
go run ./cmd/server

# Or build and run the binary
go build -o bin/server ./cmd/server
./bin/server
```

## ⚙️ Environment Variables

The application supports the following environment variables for configuration:

| Variable | Default | Description |
|----------|---------|-------------|
| `HOST` | `0.0.0.0` | Server bind address |
| `PORT` | `8080` | Server port |
| `ENV` | `development` | Application environment |
| `DEBUG` | `true` | Enable debug mode and verbose logging |

### Setting Environment Variables

**Docker Compose** (modify `docker-compose.yml`):
```yaml
services:
  app:
    environment:
      - PORT=8080
      - ENV=production
      - DEBUG=false
```

**Local Development**:
```bash
# Linux/macOS
export PORT=9000
export ENV=production
go run ./cmd/server

# Windows
set PORT=9000
set ENV=production
go run ./cmd/server
```

**Docker Run**:
```bash
docker run -e PORT=9000 -e ENV=production -p 9000:9000 acquire-app
```

## 🔧 Docker Configuration

### Multi-Stage Dockerfile

The project uses an optimized multi-stage build:

1. **Builder Stage**: Alpine Linux with Go 1.24
   - Installs build dependencies (git, ca-certificates)
   - Downloads Go modules
   - Compiles static binary with CGO disabled

2. **Runtime Stage**: Distroless Debian 11
   - Minimal attack surface (no shell, package manager)
   - Non-root user execution
   - Only essential runtime files

### Docker Compose Services

```yaml
services:
  app:
    build: .              # Uses local Dockerfile
    ports:
      - "8080:8080"        # Expose on host port 8080
    restart: unless-stopped
    environment:
      - PORT=8080
```

**Future Database Integration** (commented out):
```yaml
  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: appdb
      POSTGRES_USER: user
      POSTGRES_PASSWORD: password
    ports:
      - "5432:5432"
```

### Docker Commands

```bash
# Build and start services
docker compose up --build

# Start in detached mode
docker compose up -d --build

# View logs
docker compose logs -f app

# Stop services
docker compose down

# Rebuild after changes
docker compose up --build --force-recreate

# Remove all containers and volumes
docker compose down -v
```

## 🛠️ API Endpoints

### Current Endpoints

| Method | Path | Description | Response |
|--------|------|-------------|----------|
| `GET` | `/` | Main application interface | HTML page |
| `GET` | `/health` | Health check endpoint | JSON status |
| `GET` | `/css/*` | CSS stylesheets | Static files |
| `GET` | `/js/*` | JavaScript files | Static files |

### Health Check Response
```json
{
  "status": "ok",
  "timestamp": 1703123456
}
```

## 📋 TODO Areas for Future Development

The following areas are marked for future implementation:

### 🔌 WebUSB API Integration
**Location**: `cmd/server/main.go` (lines 53-58), `web/js/app.js` (lines 13-30)

**Planned Endpoints**:
```go
// TODO: Add WebUSB API endpoints here
// Future endpoints for WebUSB communication:
// - POST /api/webusb/connect
// - POST /api/webusb/disconnect  
// - POST /api/webusb/transfer
// - GET /api/webusb/devices
```

**Frontend Integration**:
```javascript
// TODO: Implement WebUSB requestDevice functionality
// This will handle the connection to the intra-oral capture device
// const device = await navigator.usb.requestDevice({
//     filters: [{
//         vendorId: 0x????  // Replace with actual vendor ID
//     }]
// });
```

### 🗄️ Database Integration
**Priority**: High  
**Description**: PostgreSQL integration for data persistence

**Required Changes**:
- Uncomment PostgreSQL service in `docker-compose.yml`
- Add database connection configuration
- Implement data models and migrations
- Add database connection pooling

### 🔐 Authentication System
**Priority**: Medium  
**Description**: User authentication and session management

**Components Needed**:
- JWT token authentication
- User registration/login endpoints
- Session middleware
- Role-based access control

### 🧪 Testing Suite
**Priority**: High  
**Description**: Comprehensive test coverage

**Test Types**:
- Unit tests for handlers and config
- Integration tests for API endpoints
- End-to-end tests for web interface
- Load testing for performance validation

### 📊 Logging & Monitoring
**Priority**: Medium  
**Description**: Enhanced observability

**Features**:
- Metrics collection (Prometheus)
- Distributed tracing
- Error reporting integration
- Performance monitoring dashboards

### 🔧 Configuration Management
**Priority**: Low  
**Description**: Advanced configuration options

**Enhancements**:
- Configuration file support (YAML/JSON)
- Configuration validation
- Hot-reload capabilities
- Environment-specific configs

## 🧪 Testing & Development

### Running Tests
```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

### Building for Production
```bash
# Build optimized binary
CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

# Build Docker image
docker build -t acquire-app:latest .

# Multi-platform build
docker buildx build --platform linux/amd64,linux/arm64 -t acquire-app:latest .
```

### Development Workflow

1. **Make Changes**: Edit source code
2. **Test Locally**: `go run ./cmd/server`
3. **Test with Docker**: `docker compose up --build`
4. **Run Tests**: `go test ./...`
5. **Commit Changes**: Follow conventional commits

## 🔍 Troubleshooting

### Common Issues

**Port Already in Use**:
```bash
# Check what's using port 8080
lsof -i :8080

# Kill the process or change PORT environment variable
export PORT=9000
```

**Docker Build Fails**:
```bash
# Clean Docker cache
docker system prune -a

# Rebuild without cache
docker compose build --no-cache
```

**Module Download Issues**:
```bash
# Clean module cache
go clean -modcache

# Re-download modules
go mod download
```

### Logs & Debugging

```bash
# View application logs (Docker)
docker compose logs -f app

# View system logs
journalctl -f -u docker

# Debug with verbose output
DEBUG=true go run ./cmd/server
```

## 🤝 Contributing

1. **Fork** the repository
2. **Create** a feature branch: `git checkout -b feature/amazing-feature`
3. **Make** your changes following the existing patterns
4. **Test** thoroughly with both local and Docker environments
5. **Commit** with descriptive messages: `git commit -m 'feat: add amazing feature'`
6. **Push** to your branch: `git push origin feature/amazing-feature`
7. **Create** a Pull Request with detailed description

### Code Style Guidelines

- Follow Go formatting conventions (`go fmt`)
- Write descriptive commit messages
- Add tests for new functionality
- Update documentation for API changes
- Use structured logging with `slog`

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Fiber](https://github.com/gofiber/fiber) - Fast and flexible web framework
- [Gorilla Mux](https://github.com/gorilla/mux) - Powerful HTTP router
- [Docker](https://www.docker.com/) - Containerization platform
- [Go](https://golang.org/) - The Go programming language

## 📚 File-by-File Implementation Guide

This section provides detailed best-practice explanations for each skeleton file in the project to guide future development and agent implementations.

### 🚀 Entry Points & Core Logic

#### `cmd/server/main.go` - Application Entry Point
**Purpose**: Main server entry point with Fiber v2 framework integration

**Key Implementation Details**:
- **Structured Logging**: Uses Go's `log/slog` with JSON format for production readiness
- **Graceful Shutdown**: 30-second timeout with proper signal handling (SIGTERM, SIGINT)
- **Error Handling**: Centralized error handler with structured logging
- **Middleware Stack**: Recovery middleware + request logging for debugging
- **Static File Serving**: All `/web` directory files served at root path
- **Health Check**: Simple endpoint at `/health` for monitoring/load balancers

**Extension Points**:
```go
// Lines 53-58: WebUSB API endpoints placeholder
// Add future endpoints here:
// - POST /api/webusb/connect
// - POST /api/webusb/disconnect  
// - POST /api/webusb/transfer
// - GET /api/webusb/devices
```

**Best Practices Applied**:
- Clean separation of concerns (config loading, middleware setup, routing)
- Production-ready logging with structured format
- Non-blocking server startup with goroutines
- Signal-based graceful shutdown pattern
- Environment-based configuration loading

---

### ⚙️ Configuration Management

#### `internal/config/config.go` - Environment Configuration
**Purpose**: Centralized configuration management with environment variable support

**Key Implementation Details**:
- **Go Conventions**: Follows `internal/` package convention for private code
- **Environment Variables**: Supports `HOST`, `PORT`, `ENV`, `DEBUG` with sensible defaults
- **Type Safety**: Proper boolean parsing with fallback handling
- **Zero Dependencies**: Uses only Go standard library

**Configuration Values**:
```go
type Config struct {
    Host        string // Default: "0.0.0.0" (container-friendly)
    Port        string // Default: "8080" (common web port)
    Environment string // Default: "development" (safe default)
    Debug       bool   // Default: true (development-friendly)
}
```

**Extension Guidelines**:
- Add new config fields to struct with appropriate tags
- Create helper functions for complex types (durations, URLs, etc.)
- Consider validation for critical configuration values
- Use environment variable naming convention: `APP_CONFIG_NAME`

**Best Practices Applied**:
- Fail-safe defaults for all configuration values
- Clear separation between string and boolean parsing
- No external dependencies for configuration loading
- Simple, predictable API for config consumers

---

### 🎯 Request Handling (Legacy)

#### `internal/handlers/handlers.go` - HTTP Handlers
**Purpose**: Legacy HTTP handlers (currently unused, replaced by Fiber inline handlers)

**Key Implementation Details**:
- **Standard Library**: Uses `net/http` instead of Fiber (legacy approach)
- **Structured Responses**: Proper JSON response types with timestamps
- **Error Handling**: Graceful JSON encoding error handling
- **HTML Embedding**: Direct HTML string embedding (not recommended for production)

**Current Status**: 
❗ **Note**: This file is not currently used by the main application. The main.go uses Fiber's inline handlers instead.

**Future Integration**:
If moving back to standard library HTTP or creating separate handler modules:
```go
// Register handlers with Fiber
app.Get("/health", adapters.FiberHandler(handlers.HealthHandler))
app.Get("/", adapters.FiberHandler(handlers.IndexHandler))
```

**Best Practices for Handler Development**:
- Always set proper Content-Type headers
- Use structured response types for JSON APIs
- Implement proper error handling with appropriate HTTP status codes
- Consider using handler constructors for dependency injection
- Separate business logic from HTTP handling concerns

---

### 🌐 Frontend Assets

#### `web/index.html` - Main Web Interface
**Purpose**: Primary user interface for intra-oral capture functionality

**Key Implementation Details**:
- **Semantic HTML5**: Proper document structure with `lang` attribute
- **Responsive Design**: Viewport meta tag for mobile compatibility
- **Accessibility**: Semantic elements (`header`, `main`) for screen readers
- **Clean Architecture**: Separation of structure (HTML), styling (CSS), behavior (JS)

**Interface Elements**:
- Single "Acquire" button as primary user interaction
- Header with application title
- Container-based layout for future expansion

**Extension Guidelines**:
```html
<!-- Add new UI components within the main container -->
<div class="container">
    <button id="acquire-btn" class="acquire-button">Acquire</button>
    <!-- Future: Device status indicators -->
    <!-- Future: Capture progress bar -->
    <!-- Future: Results display area -->
</div>
```

---

#### `web/css/style.css` - Discord-Dark Theme Styling
**Purpose**: Modern dark theme styling inspired by Discord's interface

**Key Implementation Details**:
- **Design System**: Consistent color palette and typography
- **Flexbox Layout**: Modern CSS layout with proper centering
- **Interactive Elements**: Hover, active, and focus states for accessibility
- **Custom Scrollbars**: Themed scrollbars matching the dark aesthetic

**Color Palette**:
```css
/* Primary Colors */
--background: #2F3136;     /* Main background */
--secondary: #202225;      /* Header background */
--accent: #5865F2;         /* Primary button color */
--text: #FFFFFF;           /* Primary text */
--border: #40444B;         /* Subtle borders */
```

**Component System**:
- `.acquire-button`: Primary action button with hover effects
- `.container`: Responsive content wrapper
- Global resets and responsive typography

**Extension Guidelines**:
- Maintain the dark theme color scheme
- Use consistent spacing units (rem-based)
- Add new components following BEM methodology
- Ensure all interactive elements have proper focus states

---

#### `web/js/app.js` - Client-Side JavaScript
**Purpose**: Application logic and future WebUSB device integration

**Key Implementation Details**:
- **DOM Ready Pattern**: Proper event listener setup after DOM load
- **Event Delegation**: Clean separation of event binding and handling
- **WebUSB Placeholder**: Structured comments showing planned implementation
- **Error Handling**: Try-catch blocks for async operations

**Current Functionality**:
```javascript
// Event binding
document.addEventListener('DOMContentLoaded', function() {
    const acquireButton = document.getElementById('acquire-btn');
    if (acquireButton) {
        acquireButton.addEventListener('click', handleAcquireClick);
    }
});
```

**WebUSB Implementation Plan** (Lines 13-30):
```javascript
// TODO: Replace with actual vendor ID and device specifications
const device = await navigator.usb.requestDevice({
    filters: [{
        vendorId: 0x????  // Replace with actual vendor ID
    }]
});
```

**Extension Guidelines**:
- Add device detection and connection status UI updates
- Implement proper error user feedback (not just console)
- Add data capture and processing logic
- Consider using modern ES modules for larger applications

---

### 🐳 Containerization

#### `Dockerfile` - Multi-Stage Container Build
**Purpose**: Optimized Docker container with security and efficiency focus

**Multi-Stage Architecture**:

**Stage 1 - Builder** (Lines 2-20):
```dockerfile
FROM golang:1.24-alpine AS builder
# - Alpine base for minimal build environment
# - Git and CA certificates for dependency management
# - Static binary compilation (CGO_ENABLED=0)
# - Security: Removes build tools from final image
```

**Stage 2 - Runtime** (Lines 23-41):
```dockerfile
FROM gcr.io/distroless/static-debian11:nonroot
# - Distroless base (no shell, package manager)
# - Non-root user execution for security
# - Minimal attack surface
# - Only essential runtime components
```

**Security Features**:
- Non-root user execution (`USER nonroot:nonroot`)
- Minimal base image (distroless)
- Static binary compilation
- CA certificates for HTTPS connectivity

**Optimization Features**:
- Multi-stage build reduces final image size
- Only production assets in final image
- Layer caching for dependency downloads

---

#### `docker-compose.yml` - Container Orchestration
**Purpose**: Development and deployment container orchestration

**Service Configuration**:
```yaml
services:
  app:
    build: .                    # Uses local Dockerfile
    ports: ["8080:8080"]        # Host:Container port mapping
    restart: unless-stopped     # Auto-restart policy
    environment:
      - PORT=8080              # Container configuration
```

**Future Database Integration** (Lines 13-23):
- PostgreSQL service ready for uncommented use
- Volume management for data persistence
- Environment-based database configuration
- Network isolation between services

**Development Workflow**:
```bash
# IMPORTANT: Always rebuild when code changes
docker compose up --build  # Rule: Container must rebuild for file changes
```

---

#### `.dockerignore` - Build Context Optimization
**Purpose**: Excludes unnecessary files from Docker build context

**Exclusion Categories**:
- **Development Files**: IDE configs, swap files, local env files
- **Documentation**: README files, markdown documentation
- **Build Artifacts**: Compiled binaries, test coverage files
- **Dependencies**: node_modules, vendor directories
- **Version Control**: .git directory and related files

**Performance Impact**:
- Reduces build context size
- Faster Docker builds
- Smaller intermediate layers
- Security: Excludes sensitive development files

---

### 📦 Go Module Management

#### `go.mod` - Module Definition
**Purpose**: Go module definition with dependency management

**Module Configuration**:
```go
module acquire-app          // Module name
go 1.24                     // Latest Go version requirement
```

**Key Dependencies**:
- **Fiber v2** (`github.com/gofiber/fiber/v2`): Modern web framework
- **Gorilla Mux** (`github.com/gorilla/mux`): HTTP router (available but unused)
- **Indirect Dependencies**: Automatically managed compression, UUID, and system libraries

**Best Practices Applied**:
- Semantic versioning for all dependencies
- Minimal direct dependencies
- Latest stable Go version requirement
- Clear module naming convention

---

### 🔄 Development Workflow Integration

**Key Integration Points**:

1. **Configuration Flow**: `config.go` → `main.go` → Environment variables
2. **Static Assets**: `web/` → Fiber static middleware → Browser
3. **Container Pipeline**: Source code → Multi-stage Dockerfile → Production image
4. **Development Loop**: Code changes → `docker compose up --build` → Testing

**Critical Dependencies**:
- All web assets must be in `/web` directory for proper serving
- Container rebuilds required for any source code changes
- Environment variables override default configuration values
- Health check endpoint required for load balancer integration

**Future Extension Areas**:
1. **WebUSB Integration**: `web/js/app.js` + `cmd/server/main.go` API endpoints
2. **Database Layer**: Uncomment PostgreSQL in `docker-compose.yml`
3. **Authentication**: Add middleware in `main.go` before route handlers
4. **Testing**: Create `*_test.go` files following Go conventions
5. **Logging**: Extend structured logging in configuration and handlers

This file structure provides a solid foundation for a production-ready Go web application with modern containerization and development practices.

---

**Last Updated**: This README reflects the current state of the project as of the latest commit. For the most up-to-date information, refer to the project documentation and source code.
