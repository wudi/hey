# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build hey-fpm binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o hey-fpm ./cmd/hey-fpm
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o hey ./cmd/hey

# Runtime stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 hey && \
    adduser -D -u 1000 -G hey hey

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /build/hey-fpm /usr/local/bin/hey-fpm
COPY --from=builder /build/hey /usr/local/bin/hey

# Create directories
RUN mkdir -p /var/www /var/run /var/log/hey-fpm && \
    chown -R hey:hey /var/www /var/run /var/log/hey-fpm

# Switch to non-root user
USER hey

# Expose FastCGI port
EXPOSE 9000

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD netstat -an | grep 9000 || exit 1

# Default command
CMD ["hey-fpm", "--nodaemonize", "--listen", "0.0.0.0:9000", "--pid", "/var/run/hey-fpm.pid"]