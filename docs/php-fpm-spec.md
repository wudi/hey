# PHP-FPM Implementation Specification

## Executive Summary

Implement a complete FastCGI Process Manager (FPM) for Hey-Codex, enabling the interpreter to handle web requests through the FastCGI protocol. This allows Hey-Codex to integrate with web servers like Nginx and Apache as a drop-in replacement for PHP-FPM.

## Architecture Overview

### Master-Worker Process Model

```
┌─────────────────────────────────────────────────┐
│              Master Process                      │
│  - Configuration Management                      │
│  - Worker Pool Management                        │
│  - Signal Handling (SIGTERM, SIGUSR1, SIGUSR2)  │
│  - Health Monitoring                             │
└─────────────────┬───────────────────────────────┘
                  │
        ┌─────────┴─────────┬─────────────┬────────
        │                   │             │
┌───────▼────────┐  ┌──────▼──────┐  ┌──▼────────┐
│ Worker Process │  │   Worker    │  │  Worker   │
│  - TCP/Unix    │  │  Process    │  │  Process  │
│    Socket      │  │             │  │           │
│  - Request     │  │             │  │           │
│    Handler     │  │             │  │           │
│  - VM Context  │  │             │  │           │
└────────────────┘  └─────────────┘  └───────────┘
```

### FastCGI Protocol Implementation

```
Web Server (Nginx)  →  FastCGI Protocol  →  Hey-FPM
                       ┌──────────────┐
                       │ FCGI_BEGIN   │
                       │ FCGI_PARAMS  │
                       │ FCGI_STDIN   │
                       └──────────────┘
                              ↓
                       ┌──────────────┐
                       │  VM Execute  │
                       └──────────────┘
                              ↓
                       ┌──────────────┐
                       │ FCGI_STDOUT  │
                       │ FCGI_STDERR  │
                       │ FCGI_END     │
                       └──────────────┘
```

## FastCGI Protocol Specification

### Record Structure

```go
type FCGIRecord struct {
    Version       uint8   // FCGI_VERSION_1
    Type          uint8   // Record type
    RequestID     uint16  // Request identifier
    ContentLength uint16  // Length of content
    PaddingLength uint8   // Length of padding
    Reserved      uint8   // Reserved byte
    Content       []byte  // Record content
    Padding       []byte  // Padding bytes
}
```

### Record Types

| Type | Value | Description |
|------|-------|-------------|
| FCGI_BEGIN_REQUEST | 1 | Begins a request |
| FCGI_ABORT_REQUEST | 2 | Aborts a request |
| FCGI_END_REQUEST | 3 | Ends a request |
| FCGI_PARAMS | 4 | Name-value pairs |
| FCGI_STDIN | 5 | Standard input stream |
| FCGI_STDOUT | 6 | Standard output stream |
| FCGI_STDERR | 7 | Standard error stream |
| FCGI_DATA | 8 | Filter data stream |
| FCGI_GET_VALUES | 9 | Query variables |
| FCGI_GET_VALUES_RESULT | 10 | Variable query result |
| FCGI_UNKNOWN_TYPE | 11 | Unknown type response |

### Begin Request Body

```go
type BeginRequestBody struct {
    Role     uint16 // FCGI_RESPONDER = 1
    Flags    uint8  // FCGI_KEEP_CONN = 1
    Reserved [5]byte
}
```

### End Request Body

```go
type EndRequestBody struct {
    AppStatus      uint32 // Application exit code
    ProtocolStatus uint8  // Protocol-level status
    Reserved       [3]byte
}
```

### Protocol Status Codes

| Status | Value | Description |
|--------|-------|-------------|
| FCGI_REQUEST_COMPLETE | 0 | Normal end of request |
| FCGI_CANT_MPX_CONN | 1 | Multiplexing not supported |
| FCGI_OVERLOADED | 2 | Too many requests |
| FCGI_UNKNOWN_ROLE | 3 | Unknown role |

## Component Architecture

### 1. FastCGI Protocol Handler (`pkg/fastcgi`)

```go
package fastcgi

// Core protocol implementation
type Protocol struct {
    conn net.Conn
}

func (p *Protocol) ReadRecord() (*Record, error)
func (p *Protocol) WriteRecord(rec *Record) error
func (p *Protocol) ReadParams() (map[string]string, error)
func (p *Protocol) WriteStdout(data []byte) error
func (p *Protocol) WriteStderr(data []byte) error
func (p *Protocol) EndRequest(appStatus uint32, protocolStatus uint8) error
```

### 2. Request Handler (`pkg/fpm/handler`)

```go
package handler

type RequestHandler struct {
    vmFactory *vmfactory.VMFactory
}

func (h *RequestHandler) HandleRequest(ctx context.Context, req *Request) error {
    // 1. Parse FastCGI params to CGI variables
    // 2. Setup VM execution context
    // 3. Capture stdout/stderr
    // 4. Execute PHP script
    // 5. Send response through FastCGI
}
```

### 3. Worker Pool Manager (`pkg/fpm/pool`)

```go
package pool

type WorkerPool struct {
    config     *PoolConfig
    workers    []*Worker
    requests   chan *Request
    quit       chan struct{}
}

type PoolConfig struct {
    ProcessManagement string // static, dynamic, ondemand
    MaxChildren       int
    StartServers      int
    MinSpareServers   int
    MaxSpareServers   int
    MaxRequests       int    // Restart worker after N requests
}

func (p *WorkerPool) Start() error
func (p *WorkerPool) Stop() error
func (p *WorkerPool) ScaleWorkers() error
```

### 4. Master Process (`pkg/fpm/master`)

```go
package master

type Master struct {
    config    *Config
    pools     []*pool.WorkerPool
    listener  net.Listener
    sigChan   chan os.Signal
}

func (m *Master) Start() error {
    // 1. Load configuration
    // 2. Setup signal handlers
    // 3. Create worker pools
    // 4. Listen on socket
    // 5. Accept connections and dispatch to workers
}

func (m *Master) handleSignals()
func (m *Master) reloadConfig()
func (m *Master) gracefulShutdown()
```

### 5. Configuration Parser (`pkg/fpm/config`)

```go
package config

type Config struct {
    Global GlobalConfig
    Pools  []PoolConfig
}

type GlobalConfig struct {
    PIDFile      string
    ErrorLog     string
    LogLevel     string
    EmergencyRestartThreshold int
    EmergencyRestartInterval  int
}

func LoadConfig(path string) (*Config, error)
```

## Configuration File Format

### php-fpm.conf

```ini
[global]
pid = /var/run/hey-fpm.pid
error_log = /var/log/hey-fpm.log
log_level = notice

[www]
listen = 127.0.0.1:9000
listen.backlog = 511
listen.owner = www-data
listen.group = www-data
listen.mode = 0660

pm = dynamic
pm.max_children = 50
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 35
pm.max_requests = 500

request_terminate_timeout = 30s
slowlog = /var/log/hey-fpm-slow.log
```

## CGI Variable Mapping

FastCGI PARAMS → PHP Superglobals

```go
func mapFastCGIParamsToVMContext(params map[string]string, vmCtx *vm.ExecutionContext) {
    // $_SERVER
    server := values.NewArray()
    for k, v := range params {
        server.ArraySet(values.NewString(k), values.NewString(v))
    }
    vmCtx.GlobalVars.Store("$_SERVER", server)

    // $_GET (parse QUERY_STRING)
    if qs, ok := params["QUERY_STRING"]; ok {
        vmCtx.GlobalVars.Store("$_GET", parseQueryString(qs))
    }

    // $_POST (parse STDIN for POST requests)
    if method, ok := params["REQUEST_METHOD"]; ok && method == "POST" {
        vmCtx.GlobalVars.Store("$_POST", parsePostData(stdin))
    }

    // $_COOKIE (parse HTTP_COOKIE)
    if cookie, ok := params["HTTP_COOKIE"]; ok {
        vmCtx.GlobalVars.Store("$_COOKIE", parseCookies(cookie))
    }
}
```

## Key FastCGI Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| SCRIPT_FILENAME | Full path to PHP script | /var/www/index.php |
| REQUEST_METHOD | HTTP method | GET, POST, PUT |
| QUERY_STRING | URL query string | foo=bar&baz=qux |
| CONTENT_TYPE | Request content type | application/x-www-form-urlencoded |
| CONTENT_LENGTH | Request body length | 1024 |
| HTTP_HOST | HTTP Host header | example.com |
| REQUEST_URI | Full request URI | /path/to/script.php?query |
| REMOTE_ADDR | Client IP address | 192.168.1.100 |
| SERVER_PROTOCOL | HTTP protocol version | HTTP/1.1 |

## Process Management Strategies

### 1. Static

```go
// Fixed number of workers, always running
pm.max_children = 50
```

### 2. Dynamic

```go
// Workers scale between min and max based on load
pm.max_children = 50
pm.start_servers = 5
pm.min_spare_servers = 5
pm.max_spare_servers = 35
```

### 3. Ondemand

```go
// Workers created on-demand, killed after idle timeout
pm.max_children = 50
pm.process_idle_timeout = 10s
```

## Signal Handling

| Signal | Action |
|--------|--------|
| SIGTERM | Graceful shutdown (wait for requests to finish) |
| SIGINT | Immediate shutdown |
| SIGUSR1 | Reopen log files |
| SIGUSR2 | Reload configuration and graceful worker restart |
| SIGQUIT | Graceful shutdown |

## Memory Management

### Request Lifecycle

```go
func (w *Worker) handleRequest(req *Request) {
    // 1. Create new VM context
    vmCtx := vm.NewExecutionContext()

    // 2. Setup CGI variables
    setupCGIVars(vmCtx, req.Params)

    // 3. Execute script
    err := w.vm.Execute(vmCtx, ...)

    // 4. Send response
    sendResponse(req.Protocol, vmCtx.OutputBuffer)

    // 5. Cleanup
    w.vm.CallAllDestructors(vmCtx)

    // 6. Check max_requests limit
    w.requestCount++
    if w.requestCount >= w.config.MaxRequests {
        w.restart()
    }
}
```

## Performance Optimizations

### 1. Connection Pooling

```go
type ConnPool struct {
    conns    chan net.Conn
    maxConns int
}
```

### 2. Bytecode Caching

```go
type OpcacheManager struct {
    cache map[string]*CompiledScript
    mu    sync.RWMutex
}

func (o *OpcacheManager) Get(file string) (*CompiledScript, bool)
func (o *OpcacheManager) Set(file string, compiled *CompiledScript)
```

### 3. Preforking Workers

```go
// Prefork workers during startup to avoid cold start latency
func (p *WorkerPool) preforkWorkers() {
    for i := 0; i < p.config.StartServers; i++ {
        p.spawnWorker()
    }
}
```

## Health Monitoring

### Status Endpoint

```go
type StatusHandler struct {
    pool *WorkerPool
}

func (s *StatusHandler) GetStatus() *PoolStatus {
    return &PoolStatus{
        Pool:             s.pool.Name,
        ProcessManager:   s.pool.Config.ProcessManagement,
        StartTime:        s.pool.StartTime,
        AcceptedConn:     s.pool.Stats.AcceptedConn,
        ListenQueue:      s.pool.Stats.ListenQueue,
        MaxListenQueue:   s.pool.Stats.MaxListenQueue,
        ActiveProcesses:  s.pool.Stats.ActiveProcesses,
        IdleProcesses:    s.pool.Stats.IdleProcesses,
        TotalProcesses:   s.pool.Stats.TotalProcesses,
    }
}
```

### Metrics Collection

```go
type PoolMetrics struct {
    AcceptedConn   uint64
    SlowRequests   uint64
    MaxChildren    int
    ListenQueue    int
    ActiveWorkers  int
    IdleWorkers    int
}
```

## Error Handling

### Error Categories

1. **Protocol Errors**: Malformed FastCGI records
2. **Application Errors**: PHP runtime errors
3. **System Errors**: Socket errors, resource exhaustion

### Error Response

```go
func sendErrorResponse(proto *fastcgi.Protocol, reqID uint16, err error) {
    // Write error to stderr
    proto.WriteStderr(reqID, []byte(err.Error()))

    // End request with error status
    proto.EndRequest(reqID, 1, fastcgi.FCGI_REQUEST_COMPLETE)
}
```

## Testing Strategy

### 1. Unit Tests

```go
func TestFastCGIProtocol(t *testing.T)
func TestRequestHandler(t *testing.T)
func TestWorkerPool(t *testing.T)
```

### 2. Integration Tests

```go
func TestEndToEndFastCGI(t *testing.T) {
    // 1. Start FPM server
    // 2. Send FastCGI request
    // 3. Verify response
}
```

### 3. Load Tests

```bash
wrk -t4 -c100 -d30s http://localhost/test.php
ab -n 10000 -c 100 http://localhost/test.php
```

### 4. Compatibility Tests

Test with real Nginx configuration:

```nginx
location ~ \.php$ {
    fastcgi_pass 127.0.0.1:9000;
    fastcgi_index index.php;
    fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    include fastcgi_params;
}
```

## Implementation Phases

### Phase 1: Core Protocol (Week 1)
- FastCGI record parsing/serialization
- Basic request/response handling
- Single-threaded responder

### Phase 2: VM Integration (Week 1)
- CGI variable mapping
- Output buffering integration
- Error handling

### Phase 3: Process Management (Week 2)
- Master process
- Static worker pool
- Signal handling

### Phase 4: Advanced Features (Week 2)
- Dynamic/ondemand process management
- Bytecode caching (opcache)
- Health monitoring

### Phase 5: Production Readiness (Week 3)
- Configuration parser
- Logging system
- Performance tuning
- Comprehensive testing

## File Structure

```
pkg/
├── fastcgi/
│   ├── protocol.go          # FastCGI protocol implementation
│   ├── record.go            # Record types and parsing
│   ├── params.go            # Parameter parsing
│   └── protocol_test.go
├── fpm/
│   ├── master/
│   │   ├── master.go        # Master process
│   │   ├── signals.go       # Signal handling
│   │   └── master_test.go
│   ├── pool/
│   │   ├── pool.go          # Worker pool management
│   │   ├── worker.go        # Worker process
│   │   ├── scaling.go       # Dynamic scaling
│   │   └── pool_test.go
│   ├── handler/
│   │   ├── handler.go       # Request handling
│   │   ├── cgi.go           # CGI variable mapping
│   │   └── handler_test.go
│   ├── config/
│   │   ├── config.go        # Configuration structures
│   │   ├── parser.go        # INI parser
│   │   └── config_test.go
│   └── opcache/
│       ├── opcache.go       # Bytecode caching
│       └── opcache_test.go
cmd/hey-fpm/
    └── main.go              # FPM entry point
```

## Command-Line Interface

```bash
# Start FPM
hey fpm

# With custom config
hey fpm --fpm-config /etc/hey/fpm.conf

# Test configuration
hey fpm --test

# Run in foreground (no daemonize)
hey fpm --nodaemonize

# Specify socket
hey fpm --listen 127.0.0.1:9000

# Show version
hey fpm --version

# Show help
hey fpm --help
```

## Nginx Integration Example

```nginx
upstream hey_backend {
    server 127.0.0.1:9000;
    keepalive 32;
}

server {
    listen 80;
    server_name example.com;
    root /var/www;

    location ~ \.php$ {
        try_files $uri =404;
        fastcgi_pass hey_backend;
        fastcgi_index index.php;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
        fastcgi_param PATH_INFO $fastcgi_path_info;
        include fastcgi_params;

        # Connection pooling
        fastcgi_keep_conn on;

        # Timeouts
        fastcgi_connect_timeout 60s;
        fastcgi_send_timeout 60s;
        fastcgi_read_timeout 60s;
    }
}
```

## Go Standard Library Usage

Leverage Go's built-in `net/http/fcgi` package:

```go
import "net/http/fcgi"

func startFPM(listener net.Listener) error {
    handler := &PHPHandler{
        vmFactory: vmfactory.NewVMFactory(...),
    }
    return fcgi.Serve(listener, handler)
}

type PHPHandler struct {
    vmFactory *vmfactory.VMFactory
}

func (h *PHPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Extract SCRIPT_FILENAME from fcgi.ProcessEnv(r)
    env := fcgi.ProcessEnv(r)
    scriptFile := env["SCRIPT_FILENAME"]

    // Setup VM context with CGI variables
    vmCtx := h.setupContext(r, env)

    // Execute PHP script
    h.executeScript(vmCtx, scriptFile, w)
}
```

## References

1. FastCGI Specification: https://fastcgi-archives.github.io/FastCGI_Specification.html
2. PHP-FPM Documentation: https://www.php.net/manual/en/install.fpm.php
3. Go fcgi package: https://pkg.go.dev/net/http/fcgi
4. Nginx FastCGI: http://nginx.org/en/docs/http/ngx_http_fastcgi_module.html

## HTTP Functions Implementation

### Implemented Functions

The following HTTP-related PHP functions are fully implemented:

1. **`header(string $header, bool $replace = true, int $response_code = 0)`**
   - Sends a raw HTTP header
   - Respects `$replace` parameter to replace existing headers
   - Optional `$response_code` to set HTTP status code
   - Automatically fails if headers already sent

2. **`header_remove(?string $name = null)`**
   - Removes a specific HTTP header
   - If `$name` is null, removes all headers

3. **`headers_list(): array`**
   - Returns array of all headers to be sent
   - Format: `["Header-Name: value", ...]`

4. **`headers_sent(&$filename = null, &$line = null): bool`**
   - Checks if HTTP headers have already been sent
   - Returns true after first output occurs
   - Populates `$filename` and `$line` with location of first output

5. **`http_response_code(?int $response_code = null): int|false`**
   - Gets or sets HTTP response status code
   - Default is 200
   - Returns previous code when setting new code

6. **`setcookie(...)`**
   - Sets a cookie with URL encoding
   - Parameters: name, value, expires, path, domain, secure, httponly
   - Adds `Set-Cookie` header

7. **`setrawcookie(...)`**
   - Sets a cookie without URL encoding
   - Same parameters as `setcookie()`

8. **`getallheaders(): array`**
   - Returns associative array of all HTTP request headers
   - Works in both CLI and FPM modes

### HTTP Context Architecture

```go
type HTTPContext struct {
    mu             sync.RWMutex
    headers        []HTTPHeader
    responseCode   int
    headersSent    bool
    headersSentAt  string
    requestHeaders map[string]string
}
```

### Automatic Headers Sent Tracking

Headers are automatically marked as "sent" when:
- Any output is written (echo, print, etc.)
- Output occurs outside of an output buffer
- First byte is written to stdout

This matches PHP's behavior where headers must be sent before any body content.

### Integration with FPM

The FPM handler extracts HTTP headers from FastCGI PARAMS and formats the response:

```go
// Extract request headers from FastCGI params (HTTP_* variables)
extractRequestHeaders(vmCtx, req.Params)

// Execute VM
vmachine.Execute(vmCtx, ...)

// Format response with HTTP headers
httpHeaders := vmCtx.HTTPContext.FormatHeadersForFastCGI()
response.WriteString(httpHeaders)
response.Write(outBuf.Bytes())
```

### Example Usage

```php
<?php
// Set custom headers
header("Content-Type: application/json");
header("X-Custom-Header: value");

// Set response code
http_response_code(201);

// Set cookie
setcookie("session_id", "abc123", time() + 3600, "/");

// Get all headers
$headers = headers_list();
var_dump($headers);

// Check if headers sent
if (!headers_sent()) {
    header("Last-Modified: " . gmdate('D, d M Y H:i:s') . ' GMT');
}

// Output content (headers automatically sent now)
echo json_encode(["status" => "success"]);
?>
```

### Testing

Integration tests validate all HTTP functions:
- Headers are collected before output
- `headers_sent()` returns false initially
- `headers_sent()` returns true after output
- Headers cannot be modified after output starts
- Output buffering delays headers being sent
- Response codes are tracked correctly

See `/test/http_functions_test.php` for comprehensive test coverage.

## Success Criteria

✅ Successfully handles FastCGI requests from Nginx
✅ Executes PHP scripts with correct superglobal population
✅ Supports all three process management modes
✅ Handles graceful restart without dropping requests
✅ Passes load testing with 1000+ req/s
✅ Compatible with standard Nginx/Apache FastCGI configurations
✅ Implements opcache for performance
✅ Provides health monitoring endpoint
✅ Full HTTP header manipulation support
✅ Automatic headers_sent tracking on output
✅ Cookie setting functions (setcookie, setrawcookie)