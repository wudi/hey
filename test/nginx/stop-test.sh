#!/bin/bash
# Stop hey-fpm and Nginx test services

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "=== Stopping Hey-Codex FPM + Nginx Test Services ==="
echo ""

# Stop Nginx
echo "Stopping Nginx..."
if [ -f "$SCRIPT_DIR/logs/nginx.pid" ]; then
    nginx -s stop -c "$SCRIPT_DIR/nginx.conf" -p "$SCRIPT_DIR" 2>/dev/null || true
    echo "✓ Nginx stopped"
else
    echo "Nginx PID file not found, attempting graceful stop..."
    nginx -s stop -c "$SCRIPT_DIR/nginx.conf" -p "$SCRIPT_DIR" 2>/dev/null || echo "Nginx may not be running"
fi

# Stop hey-fpm
echo "Stopping hey-fpm..."
if [ -f "$SCRIPT_DIR/hey-fpm.pid" ]; then
    FPM_PID=$(cat "$SCRIPT_DIR/hey-fpm.pid")
    if kill -0 $FPM_PID 2>/dev/null; then
        kill $FPM_PID
        echo "✓ hey-fpm stopped (PID: $FPM_PID)"
    else
        echo "hey-fpm process not found (PID: $FPM_PID)"
    fi
    rm "$SCRIPT_DIR/hey-fpm.pid"
else
    echo "hey-fpm PID file not found"
    # Try to find and kill any hey-fpm process
    pkill -f hey-fpm || echo "No hey-fpm processes found"
fi

echo ""
echo "=== Services Stopped ==="