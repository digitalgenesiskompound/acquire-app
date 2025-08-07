# Build stage
FROM golang:1.24-alpine AS builder

# Install git and ca-certificates for dependencies
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Runtime stage using distroless
FROM gcr.io/distroless/static-debian11:nonroot

# Copy ca-certificates from builder stage
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary from builder stage
COPY --from=builder /app/main /

# Copy web assets
COPY --from=builder /app/web /web

# Expose port 8080
EXPOSE 8080

# Use nonroot user
USER nonroot:nonroot

# Run the binary
ENTRYPOINT ["/main"]
