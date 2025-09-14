# Build stage
FROM golang:1.21-alpine AS builder

# Install git and ca-certificates (needed for go modules)
RUN apk add --no-cache git ca-certificates tzdata

# Create a non-root user
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /build

# Copy go modules files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o index-explorer \
    ./cmd/

# Runtime stage
FROM alpine:3.18

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata curl

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Copy user information
COPY --from=builder /etc/passwd /etc/passwd

# Create necessary directories
RUN mkdir -p /app/configs /app/web /app/logs

# Copy binary and configuration files
COPY --from=builder /build/index-explorer /app/
COPY --from=builder /build/configs/ /app/configs/
COPY --from=builder /build/web/ /app/web/

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Set working directory
WORKDIR /app

# Expose port
EXPOSE 8082

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8082/health || exit 1

# Set environment variables
ENV GIN_MODE=release
ENV CONFIG_PATH=configs/config.yaml

# Labels for better container management
LABEL maintainer="Elasticsearch Playground Team"
LABEL description="Write-optimized Elasticsearch Index & Document Explorer"
LABEL version="1.0.0"
LABEL org.opencontainers.image.title="Index Explorer"
LABEL org.opencontainers.image.description="High-performance write-optimized Elasticsearch operations"
LABEL org.opencontainers.image.source="https://github.com/saif-islam/es-playground"
LABEL org.opencontainers.image.vendor="ES Playground"

# Run the binary
CMD ["./index-explorer"]