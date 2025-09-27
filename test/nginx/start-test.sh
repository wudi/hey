#!/bin/bash
# Start hey-fpm and Nginx for testing

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"

echo "=== Hey-Codex FPM + Nginx Test Setup ==="
echo ""

# Check if hey-fpm binary exists
if [ ! -f "$ROOT_DIR/build/hey-fpm" ]; then
    echo "Error: hey-fpm binary not found at $ROOT_DIR/build/hey-fpm"
    echo "Please run: make build-all"
    exit 1
fi

# Check if nginx is installed
if ! command -v nginx &> /dev/null; then
    echo "Error: nginx is not installed"
    echo "Please install nginx:"
    echo "  Ubuntu/Debian: sudo apt-get install nginx"
    echo "  macOS: brew install nginx"
    exit 1
fi

# Create logs directory if it doesn't exist
mkdir -p "$SCRIPT_DIR/logs"

# Check if hey-fpm is already running
if lsof -i :9000 &> /dev/null; then
    echo "Warning: Port 9000 is already in use"
    echo "Please stop any existing FPM process or change the port"
    exit 1
fi

# Check if nginx is already running on port 8080
if lsof -i :8080 &> /dev/null; then
    echo "Warning: Port 8080 is already in use"
    echo "Please stop any existing web server or change the port in nginx.conf"
    exit 1
fi

# Start hey-fpm in background with custom PID path
echo "Starting hey-fpm on 127.0.0.1:9000..."
"$ROOT_DIR/build/hey-fpm" --nodaemonize --pid "$SCRIPT_DIR/hey-fpm.pid" > "$SCRIPT_DIR/logs/fpm.log" 2>&1 &
FPM_PID=$!
echo "hey-fpm started with PID: $FPM_PID"

# Wait for FPM to start
echo "Waiting for hey-fpm to be ready..."
sleep 2

# Verify FPM is listening
if ! lsof -i :9000 &> /dev/null; then
    echo "Error: hey-fpm failed to start on port 9000"
    exit 1
fi
echo "✓ hey-fpm is listening on port 9000"

# Start Nginx
echo ""
echo "Starting Nginx on port 8080..."
nginx -c "$SCRIPT_DIR/nginx.conf" -p "$SCRIPT_DIR"

if [ $? -eq 0 ]; then
    echo "✓ Nginx started successfully"
else
    echo "Error: Failed to start Nginx"
    kill $FPM_PID
    exit 1
fi

echo ""
echo "=== Services Started Successfully ==="
echo ""
echo "hey-fpm:   127.0.0.1:9000 (PID: $FPM_PID)"
echo "Nginx:     http://localhost:8080/"
echo ""
echo "Test URLs:"
echo "  - http://localhost:8080/            (Main test page)"
echo "  - http://localhost:8080/info.php    (PHP info)"
echo "  - http://localhost:8080/headers.php (Headers test)"
echo "  - http://localhost:8080/cookies.php (Cookies test)"
echo "  - http://localhost:8080/json.php    (JSON API test)"
echo ""
echo "Logs:"
echo "  - Access: $SCRIPT_DIR/logs/access.log"
echo "  - Error:  $SCRIPT_DIR/logs/error.log"
echo "  - FPM:    $SCRIPT_DIR/logs/fpm.log"
echo ""
echo "To stop services, run: $SCRIPT_DIR/stop-test.sh"
echo "To run tests, run: $SCRIPT_DIR/run-tests.sh"
echo ""