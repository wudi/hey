# Nginx Integration Testing for Hey-Codex FPM

This directory contains configuration and test files for testing hey-fpm with Nginx.

## Prerequisites

1. Install Nginx (if not already installed):
```bash
# Ubuntu/Debian
sudo apt-get install nginx

# macOS
brew install nginx
```

2. Build hey-fpm:
```bash
cd /home/ubuntu/hey-codex
make build-all
# or
go build -o build/hey-fpm ./cmd/hey-fpm
```

## Quick Start

### 1. Start hey-fpm

```bash
# Start hey-fpm on default port 9000
./build/hey-fpm
```

The FPM server will start and listen on `127.0.0.1:9000` by default.

### 2. Start Nginx with test configuration

```bash
# Start Nginx with the test configuration
sudo nginx -c $(pwd)/test/nginx/nginx.conf -p $(pwd)/test/nginx
```

Or if you have permissions:
```bash
nginx -c $(pwd)/test/nginx/nginx.conf -p $(pwd)/test/nginx
```

### 3. Test the setup

Open your browser and visit:
- http://localhost:8080/ - Main test page
- http://localhost:8080/info.php - PHP info
- http://localhost:8080/headers.php - HTTP headers test
- http://localhost:8080/cookies.php - Cookie functionality test
- http://localhost:8080/json.php - JSON API test

Or use curl:
```bash
curl http://localhost:8080/
curl http://localhost:8080/json.php
curl -H "X-Custom: Test" http://localhost:8080/headers.php
```

## Configuration Details

### Nginx Configuration

- **Listen Port**: 8080 (to avoid conflicts with default Nginx)
- **Document Root**: `test/nginx/www/`
- **FastCGI Backend**: 127.0.0.1:9000 (hey-fpm)
- **Logs**: `test/nginx/logs/`

### Hey-FPM Configuration

Default configuration:
- **Listen**: 127.0.0.1:9000
- **Process Management**: dynamic
- **Max Workers**: Based on CPU count

## Test Pages

### 1. index.php
Main test page showing:
- PHP version and server info
- Request details (method, URI, remote address)
- HTTP headers received
- Server variables
- Links to other test pages

### 2. info.php
Displays full PHP information (phpinfo equivalent for hey-codex).

### 3. headers.php
Tests HTTP header functionality:
- Custom headers sent via `header()`
- Response code via `http_response_code()`
- Headers list via `headers_list()`
- Headers sent status via `headers_sent()`
- Request headers via `getallheaders()`

### 4. cookies.php
Tests cookie functionality:
- Setting cookies via `setcookie()`
- Reading cookies from `$_COOKIE`
- Multiple cookies with different parameters

### 5. json.php
JSON API endpoint returning:
- Request information
- HTTP headers
- Timestamp and server info

## Stopping Services

### Stop Nginx
```bash
# Using the test configuration
sudo nginx -s stop -c $(pwd)/test/nginx/nginx.conf -p $(pwd)/test/nginx
```

### Stop hey-fpm
```bash
# Find the process
ps aux | grep hey-fpm

# Kill it
kill <PID>

# Or use Ctrl+C if running in foreground
```

## Troubleshooting

### Port 8080 already in use
Change the port in `nginx.conf`:
```nginx
listen 8081;  # or any other available port
```

### hey-fpm not receiving requests
1. Check if hey-fpm is running:
```bash
ps aux | grep hey-fpm
```

2. Check if port 9000 is listening:
```bash
netstat -an | grep 9000
# or
lsof -i :9000
```

3. Check Nginx error logs:
```bash
tail -f test/nginx/logs/error.log
```

### Permission denied errors
Make sure:
- Nginx has permission to access `test/nginx/www/` directory
- hey-fpm has permission to read PHP files

### FastCGI connection refused
1. Verify hey-fpm is running on port 9000
2. Check firewall settings
3. Verify Nginx can connect to 127.0.0.1:9000

## Performance Testing

### Using Apache Bench
```bash
# Test main page with 100 requests, 10 concurrent
ab -n 100 -c 10 http://localhost:8080/

# Test JSON endpoint
ab -n 1000 -c 50 http://localhost:8080/json.php
```

### Using wrk
```bash
# Test with 4 threads, 100 connections for 30 seconds
wrk -t4 -c100 -d30s http://localhost:8080/

# Test JSON endpoint
wrk -t4 -c100 -d30s http://localhost:8080/json.php
```

### Using curl for load testing
```bash
# Simple loop test
for i in {1..100}; do
    curl -s http://localhost:8080/ > /dev/null
    echo "Request $i completed"
done
```

## Development

### Adding New Test Pages
1. Create a new PHP file in `test/nginx/www/`
2. Access it via http://localhost:8080/your-file.php

### Modifying Nginx Configuration
1. Edit `test/nginx/nginx.conf`
2. Reload Nginx:
```bash
sudo nginx -s reload -c $(pwd)/test/nginx/nginx.conf -p $(pwd)/test/nginx
```

### Viewing Logs
```bash
# Access log (all requests)
tail -f test/nginx/logs/access.log

# Error log
tail -f test/nginx/logs/error.log
```

## Expected Results

When everything is working correctly:

1. **Main page** (index.php) should display:
   - "Hey-Codex FPM is working!" message
   - Server information
   - Request details
   - HTTP headers

2. **Headers test** (headers.php) should show:
   - Custom X-Custom-Header and X-Powered-By headers
   - Response code 200
   - Request headers from browser

3. **Cookie test** (cookies.php) should:
   - Set cookies on first visit
   - Display cookies on subsequent visits
   - Show Set-Cookie headers

4. **JSON endpoint** (json.php) should return:
   - Valid JSON response
   - Content-Type: application/json header
   - Request and server information

## Integration with CI/CD

You can automate testing with a script:

```bash
#!/bin/bash
# start-test.sh

# Start hey-fpm in background
./build/hey-fpm &
FPM_PID=$!

# Wait for FPM to start
sleep 2

# Start Nginx
nginx -c $(pwd)/test/nginx/nginx.conf -p $(pwd)/test/nginx

# Run tests
curl -f http://localhost:8080/ || exit 1
curl -f http://localhost:8080/json.php || exit 1
curl -f http://localhost:8080/headers.php || exit 1

# Cleanup
nginx -s stop -c $(pwd)/test/nginx/nginx.conf -p $(pwd)/test/nginx
kill $FPM_PID

echo "All tests passed!"
```

## References

- [Nginx FastCGI Documentation](http://nginx.org/en/docs/http/ngx_http_fastcgi_module.html)
- [FastCGI Specification](https://fastcgi-archives.github.io/FastCGI_Specification.html)
- [PHP-FPM Configuration](https://www.php.net/manual/en/install.fpm.configuration.php)