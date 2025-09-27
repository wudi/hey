#!/bin/bash

set -e

echo "=== Hey-FPM Integration Test ==="
echo

echo "Step 1: Creating test PHP script..."
mkdir -p /tmp/hey-fpm-test
cat > /tmp/hey-fpm-test/info.php <<'EOF'
<?php
echo "Hey-FPM is working!\n";
echo "PHP Version: " . phpversion() . "\n";
echo "Request Method: " . $_SERVER['REQUEST_METHOD'] . "\n";
echo "Query String: " . ($_SERVER['QUERY_STRING'] ?? 'none') . "\n";
phpinfo();
?>
EOF

cat > /tmp/hey-fpm-test/test.php <<'EOF'
<?php
echo "Content-Type: text/plain\r\n\r\n";
echo "Hello from Hey-FPM!\n";
echo "Request Method: " . $_SERVER['REQUEST_METHOD'] . "\n";
echo "Script Filename: " . $_SERVER['SCRIPT_FILENAME'] . "\n";

if (!empty($_GET)) {
    echo "GET parameters:\n";
    foreach ($_GET as $key => $value) {
        echo "  $key = $value\n";
    }
}
?>
EOF

echo "Created test scripts:"
echo "  - /tmp/hey-fpm-test/test.php"
echo "  - /tmp/hey-fpm-test/info.php"
echo

echo "Step 2: Starting Hey-FPM on port 9000..."
./build/hey-fpm --listen 127.0.0.1:9000 --nodaemonize &
FPM_PID=$!

sleep 2

if ! kill -0 $FPM_PID 2>/dev/null; then
    echo "ERROR: Hey-FPM failed to start"
    exit 1
fi

echo "Hey-FPM started with PID $FPM_PID"
echo

echo "Step 3: Testing FastCGI connection with cgi-fcgi..."
if command -v cgi-fcgi &> /dev/null; then
    echo "Using cgi-fcgi to test:"
    SCRIPT_FILENAME=/tmp/hey-fpm-test/test.php \
    REQUEST_METHOD=GET \
    QUERY_STRING="name=test&foo=bar" \
    cgi-fcgi -bind -connect 127.0.0.1:9000 || true
else
    echo "cgi-fcgi not found, skipping this test"
fi

echo

echo "Step 4: Creating minimal Nginx config..."
cat > /tmp/hey-fpm-test/nginx.conf <<'EOF'
worker_processes 1;
error_log /tmp/hey-fpm-test/nginx-error.log;
pid /tmp/hey-fpm-test/nginx.pid;

events {
    worker_connections 1024;
}

http {
    access_log /tmp/hey-fpm-test/nginx-access.log;

    server {
        listen 8080;
        server_name localhost;
        root /tmp/hey-fpm-test;

        location ~ \.php$ {
            fastcgi_pass 127.0.0.1:9000;
            fastcgi_index index.php;
            fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
            fastcgi_param REQUEST_METHOD $request_method;
            fastcgi_param QUERY_STRING $query_string;
            fastcgi_param CONTENT_TYPE $content_type;
            fastcgi_param CONTENT_LENGTH $content_length;
            include /etc/nginx/fastcgi_params;
        }
    }
}
EOF

echo "Nginx config created at /tmp/hey-fpm-test/nginx.conf"
echo

if command -v nginx &> /dev/null; then
    echo "Step 5: Starting Nginx..."
    nginx -c /tmp/hey-fpm-test/nginx.conf -p /tmp/hey-fpm-test &
    NGINX_PID=$!
    sleep 1

    if ! kill -0 $NGINX_PID 2>/dev/null; then
        echo "ERROR: Nginx failed to start"
    else
        echo "Nginx started with PID $NGINX_PID"
        echo

        echo "Step 6: Testing HTTP requests..."
        echo "--- Test 1: Simple request ---"
        curl -s http://localhost:8080/test.php
        echo
        echo

        echo "--- Test 2: Request with query string ---"
        curl -s "http://localhost:8080/test.php?name=alice&age=25"
        echo
        echo

        echo "--- Test 3: phpinfo() ---"
        curl -s http://localhost:8080/info.php | head -20
        echo

        echo "Stopping Nginx..."
        kill $NGINX_PID 2>/dev/null || true
    fi
else
    echo "Nginx not found, skipping HTTP tests"
fi

echo

echo "Cleanup..."
kill $FPM_PID 2>/dev/null || true
sleep 1

echo
echo "=== Test Complete ==="
echo "To run Hey-FPM manually:"
echo "  ./build/hey-fpm --listen 127.0.0.1:9000"
echo
echo "To test with curl (requires Nginx):"
echo "  curl http://localhost:8080/test.php"