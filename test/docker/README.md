# Docker Deployment for Hey-Codex FPM

Complete Docker setup for running Hey-Codex FPM with Nginx in containers.

## Quick Start

```bash
# Build and start containers
docker compose up -d

# View logs
docker compose logs -f

# Stop containers
docker compose down
```

## Architecture

```
┌─────────────────┐         ┌─────────────────┐
│                 │         │                 │
│  Nginx          │────────▶│  Hey-FPM        │
│  (Port 8080)    │ FastCGI │  (Port 9000)    │
│                 │         │  5 Workers      │
└─────────────────┘         └─────────────────┘
        │                           │
        │                           │
    Volumes                     Volumes
 /var/www (ro)              /var/www (ro)
 nginx-logs                 hey-fpm-logs
```

## Components

### 1. hey-fpm Container
- **Base Image**: Alpine Linux (multi-stage build from golang:1.24-alpine)
- **Binary**: Statically compiled hey-fpm (CGO_ENABLED=0)
- **User**: Non-root user `hey` (UID 1000)
- **Port**: 9000 (FastCGI)
- **Process Management**: Dynamic (5 workers)
- **Health Check**: Checks port 9000 every 30s

### 2. nginx Container
- **Base Image**: nginx:alpine (official)
- **Port**: 8080 (mapped to host)
- **Configuration**: Custom nginx.conf for FastCGI
- **Backend**: hey-fpm:9000

### 3. Volumes
- `hey-fpm-logs`: Persistent logs from hey-fpm
- `nginx-logs`: Persistent logs from Nginx
- `./test/docker/www`: PHP files (read-only mount)

### 4. Network
- `hey-network`: Bridge network for inter-container communication

## Files

### Dockerfile
Multi-stage build:
1. **Builder stage**: Compiles hey-fpm and hey binaries
2. **Runtime stage**: Minimal Alpine image with binaries

Size comparison:
- Builder image: ~1.2GB (with Go toolchain)
- Runtime image: ~25MB (Alpine + binaries)

### docker-compose.yml
Defines:
- Services (hey-fpm, nginx)
- Networks
- Volumes
- Environment variables
- Health checks

### test/docker/nginx.conf
Nginx configuration optimized for FastCGI:
- Upstream backend (hey-fpm:9000)
- FastCGI parameter passing
- Connection pooling (keepalive 32)
- Buffer settings
- Static file handling

## Usage

### Start Services

```bash
# Build images and start containers
docker compose up -d

# View startup logs
docker compose logs hey-fpm
```

### Test Endpoints

```bash
# Main page
curl http://localhost:8080/

# JSON API
curl http://localhost:8080/json.php

# Headers test
curl -I http://localhost:8080/headers.php

# Cookies test
curl http://localhost:8080/cookies.php
```

### Automated Testing

```bash
# Run complete test suite
./test/docker/test.sh
```

Test includes:
- Container connectivity
- Main page rendering
- JSON API functionality
- HTTP headers
- Cookie handling
- 404 error handling

### View Logs

```bash
# Follow hey-fpm logs
docker compose logs -f hey-fpm

# Follow nginx logs
docker compose logs -f nginx

# View all logs
docker compose logs -f
```

### Access Container Shell

```bash
# hey-fpm container
docker compose exec hey-fpm sh

# nginx container
docker compose exec nginx sh
```

### Container Stats

```bash
# View resource usage
docker stats hey-fpm hey-nginx

# Or with docker compose
docker compose stats
```

### Stop Services

```bash
# Stop containers (keep volumes)
docker compose stop

# Stop and remove containers
docker compose down

# Stop and remove everything including volumes
docker compose down -v
```

## Configuration

### Customize Worker Count

Edit `docker-compose.yml`:

```yaml
command: >
  hey-fpm
  --nodaemonize
  --listen 0.0.0.0:9000
  --pid /tmp/hey-fpm.pid
  --pm dynamic
  --pm-max-children 100      # Maximum workers
  --pm-start-servers 10      # Initial workers
  --pm-min-spare-servers 5   # Minimum idle
  --pm-max-spare-servers 50  # Maximum idle
```

### Change Ports

Edit `docker-compose.yml`:

```yaml
nginx:
  ports:
    - "8081:80"  # Change host port to 8081
```

### Add PHP Files

Place PHP files in `test/docker/www/`:

```bash
cp myapp.php test/docker/www/
docker compose restart nginx
```

Files are mounted read-only, so changes require container restart.

### Custom Nginx Configuration

Edit `test/docker/nginx.conf` and restart:

```bash
vim test/docker/nginx.conf
docker compose restart nginx
```

## Production Considerations

### Security

1. **Run as non-root**: ✅ Already configured (user `hey`)
2. **Read-only filesystem**: Consider adding:
```yaml
    read_only: true
    tmpfs:
      - /tmp
      - /var/run
```

3. **Resource limits**: Add to docker-compose.yml:
```yaml
    deploy:
      resources:
        limits:
          cpus: '2.0'
          memory: 1G
        reservations:
          cpus: '1.0'
          memory: 512M
```

### Performance

1. **Worker tuning**: Adjust based on CPU cores
```bash
--pm-max-children $((CPU_CORES * 2 + 1))
```

2. **Nginx worker processes**: Set in nginx.conf:
```nginx
worker_processes auto;
```

3. **Connection pooling**: Already enabled with `keepalive 32`

### Monitoring

1. **Health checks**: Already configured for hey-fpm
2. **Logs**: Persisted in volumes
3. **Metrics**: Consider adding Prometheus exporter

### High Availability

1. **Scale workers**:
```bash
docker compose up -d --scale hey-fpm=3
```

2. **Load balancer**: Use Nginx upstream with multiple backends

3. **Docker Swarm/Kubernetes**: For orchestration

## Troubleshooting

### Container won't start

```bash
# Check logs
docker compose logs hey-fpm

# Common issues:
# - Port 9000 already in use
# - Permission issues
# - Out of memory
```

### Cannot connect to backend

```bash
# Check if hey-fpm is listening
docker compose exec hey-fpm netstat -an | grep 9000

# Check network connectivity
docker compose exec nginx ping hey-fpm

# Verify nginx configuration
docker compose exec nginx nginx -t
```

### 502 Bad Gateway

Usually means:
- hey-fpm crashed (check logs)
- hey-fpm not listening on correct port
- Network connectivity issue

```bash
# Restart hey-fpm
docker compose restart hey-fpm

# Check if port is accessible
docker compose exec nginx telnet hey-fpm 9000
```

### High memory usage

```bash
# Check stats
docker stats --no-stream

# Reduce worker count
# Edit docker-compose.yml: --pm-max-children 25
docker compose up -d
```

### Slow response times

```bash
# Check worker stats
docker compose logs hey-fpm | grep -i worker

# Increase workers if all busy
# Check nginx buffer settings
```

## Development Workflow

### 1. Modify Code

```bash
# Edit source code
vim runtime/http_functions.go

# Rebuild container
docker compose build hey-fpm

# Restart service
docker compose up -d hey-fpm
```

### 2. Add Test PHP Files

```bash
cp test.php test/docker/www/
# No restart needed if volume mounted correctly
```

### 3. Debug

```bash
# Access hey-fpm container
docker compose exec hey-fpm sh

# Check what's running
ps aux

# Test hey binary directly
/usr/local/bin/hey -r 'echo "test";'
```

## Test Results

### Automated Test Suite

```
=== Test Summary ===
Passed: 11
Failed: 0
Total:  11
```

All tests passing:
- ✅ Container connectivity
- ✅ Main page content
- ✅ PHP version display
- ✅ JSON endpoint
- ✅ JSON content type
- ✅ JSON response structure
- ✅ Headers page
- ✅ Custom headers (X-Custom-Header)
- ✅ Powered-by header (X-Powered-By)
- ✅ Cookies functionality
- ✅ 404 error handling

### Resource Usage

```
Container: hey-fpm
- CPU: ~0.01%
- Memory: ~3MB
- Workers: 5 active

Container: nginx
- CPU: ~0.00%
- Memory: ~3MB
```

### Performance

Response times (measured with curl):
- Static HTML: <10ms
- JSON API: <10ms
- PHP processing: <15ms

## References

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Nginx FastCGI Module](http://nginx.org/en/docs/http/ngx_http_fastcgi_module.html)
- [Alpine Linux](https://alpinelinux.org/)

## License

Same as Hey-Codex project.