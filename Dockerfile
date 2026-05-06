# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -extldflags '-static'" \
    -o /mcp-calculator \
    ./cmd/mcp-calculator

# Runtime stage
FROM scratch

# Import CA certificates and timezone data
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy binary
COPY --from=builder /mcp-calculator /mcp-calculator

# Expose default port
EXPOSE 8080

# Set environment defaults
ENV MCP_HOST=0.0.0.0
ENV MCP_PORT=8080
ENV MCP_LOG_LEVEL=info
ENV MCP_SESSION_TIMEOUT=10m
ENV MCP_MAX_SESSIONS=10000
ENV MCP_RATE_LIMIT=60
ENV MCP_RATE_LIMIT_BURST=10

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD ["/mcp-calculator", "-version"]

# Run as non-root (UID 65534 is nobody)
USER 65534:65534

# Entry point
ENTRYPOINT ["/mcp-calculator"]
