# Simple HTTP Server in Go

A lightweight HTTP server implementation built from scratch in Go as part of the [CodeCrafters](https://codecrafters.io) "Build Your Own HTTP Server" challenge. This project serves as Go training by creating a custom web server with advanced features like TLS support, middleware chains, and concurrent connection handling.

## 🚀 Features

- **HTTP/1.1 & HTTPS Protocol Support** - Handles basic HTTP requests and responses with full TLS encryption
- **Concurrent Connection Handling** - Advanced worker pool architecture with buffered channels for optimal performance
- **Middleware System** - Extensible middleware chain for request/response processing
- **File Operations** - Upload and download files via HTTP endpoints with security validation
- **Gzip Compression** - Automatic response compression when supported by client
- **Health Monitoring** - Real-time server metrics and health status endpoint
- **Echo Endpoint** - Simple endpoint for testing and debugging
- **User-Agent Detection** - Endpoint to retrieve client user-agent information
- **Persistent Connections** - Support for keep-alive connections with configurable timeouts
- **Command-Line Configuration** - Configurable file directory via flags
- **Graceful Shutdown** - Proper server shutdown handling with active connection cleanup
- **Request Logging** - Comprehensive request logging with timing information

## 📋 Supported Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | Returns 200 OK (health check) |
| `GET` | `/echo/{message}` | Returns the message in the response body |
| `GET` | `/user-agent` | Returns the client's User-Agent header |
| `GET` | `/files/{filename}` | Downloads a file from the server |
| `POST` | `/files/{filename}` | Uploads a file to the server |
| `GET` | `/health` | Returns server health metrics in JSON format |

## 🛠️ Installation & Usage

### Prerequisites
- Go 1.24.0 or higher

### Clone and Run
```bash
git clone <repository-url>
cd http-server-starter-go
go run main.go
```

### Command Line Options
```bash
go run main.go -directory /path/to/files
```

**Options:**
- `-directory`: Specifies the directory where files are stored (default: `/tmp/`)

**Default Ports:**
- HTTP: `4221`
- HTTPS: `4222`

## 📡 API Examples

### Basic Health Check
```bash
curl http://localhost:4221/
# Response: 200 OK
```

### Server Health Metrics
```bash
curl http://localhost:4221/health
# Response: 
# {
#   "status": "healthy",
#   "timestamp": "2025-08-08T15:04:05Z",
#   "uptime": "1h23m45s",
#   "active_connections": 5,
#   "total_requests": 1247
# }
```

### Echo Endpoint
```bash
curl http://localhost:4221/echo/hello-world
# Response: hello-world
```

### User Agent Detection
```bash
curl http://localhost:4221/user-agent
# Response: curl/7.81.0
```

### File Upload
```bash
curl -X POST -d "file content here" http://localhost:4221/files/example.txt
# Response: 201 Created
```

### File Download
```bash
curl http://localhost:4221/files/example.txt
# Response: file content here
```

### Gzip Compression
```bash
curl -H "Accept-Encoding: gzip" http://localhost:4221/echo/compress-this-text
# Response: compressed content with Content-Encoding: gzip header
```

### HTTPS/TLS Support
```bash
curl -k https://localhost:4222/health
# Response: JSON health metrics over encrypted connection
```

## 🏗️ Architecture

### Core Components

#### `server` Package
- **Server Manager**: Manages both HTTP and HTTPS TCP servers with graceful shutdown
- **Worker Pool**: Handles connections using a worker pool pattern with configurable concurrency
- **Connection Tracking**: Tracks open connections and request metrics for monitoring
- **Signal Handling**: Listens for OS signals to trigger graceful shutdown

#### `http` Package
- **Request Parser**: Parses incoming HTTP requests into structured data with header validation
- **Response Builder**: Constructs HTTP responses with proper headers and status codes
- **Connection Management**: Handles persistent connections with configurable timeouts

#### `router` Package
- **HTTP Router**: Dispatches requests to appropriate handlers with middleware support
- **Pattern Matching**: Support for prefix-based route matching
- **Middleware Integration**: Seamless middleware chain execution

#### `handler` Package
- **Handler Interface**: Common interface for all request handlers with function adapter
- **EchoHandler**: Handles `/echo/*` endpoints with path parameter extraction
- **FileHandler**: Handles file operations (`/files/*`) with security validation
- **UserAgentHandler**: Handles `/user-agent` endpoint
- **HealthHandler**: Provides server metrics and health status

#### `middleware` Package
- **Middleware Chain**: Composable middleware system for cross-cutting concerns
- **Gzip Compression**: Automatic response compression based on Accept-Encoding headers
- **Request Logging**: Comprehensive logging with timing and status code information

#### `config` Package
- **Configuration Management**: Server constants and configuration with TLS certificates
- **TLS Setup**: Built-in TLS configuration with self-signed certificates for development

### Request Flow
1. TCP connection established (HTTP or HTTPS)
2. Connection queued in buffered channel for worker pool
3. Worker goroutine picks up connection
4. Request routed through middleware chain
5. Request dispatched to appropriate handler
6. HTTP request parsed from raw bytes
7. Response generated with proper headers and status codes
9. Response sent back to client
10. Connection maintained for keep-alive or closed based on headers

### Worker Pool Architecture
- Configurable number of worker goroutines (default: 10)
- Buffered connection channel prevents blocking on Accept()
- Graceful connection rejection when worker pool is saturated
- Atomic counters for connection and request tracking

## 🔧 Technical Details

### HTTP Features Implemented
- ✅ HTTP/1.1 protocol parsing with complete header support
- ✅ HTTPS/TLS encryption with configurable certificates
- ✅ Request method handling (GET, POST) with extensible architecture
- ✅ Header parsing and validation with whitespace trimming
- ✅ Request body handling with Content-Length validation
- ✅ Status code responses (200, 201, 400, 404, 500) with proper messages
- ✅ Content-Type and Content-Length headers with automatic calculation
- ✅ Connection management (keep-alive/close) with timeout handling
- ✅ Gzip compression support
- ✅ Middleware system for extensible request processing

### Concurrency
- Worker pool pattern with configurable goroutine count
- Buffered channels prevent connection blocking
- Atomic operations for thread-safe metrics tracking
- Graceful connection cleanup with WaitGroup synchronization
- Timeout-based connection management

### Error Handling
- Comprehensive error responses for malformed requests
- File operation error handling with appropriate HTTP status codes
- Connection error recovery with proper cleanup
- Directory traversal protection in file handlers
- Graceful degradation for worker pool saturation

### Security
- Directory traversal protection preventing `../` attacks
- Content-Length header validation preventing buffer overflows
- Safe file path handling with filepath.Join
- Input validation for file names and headers
- TLS encryption for secure communication

### Performance
- Connection pooling with configurable worker count
- Keep-alive connections reduce overhead
- Gzip compression reduces bandwidth usage
- Atomic operations minimize lock contention
- Buffered channels optimize connection handling

## 🧪 Testing

The project includes comprehensive unit tests for core components:

### HTTP Tests
```bash
go test ./http/...
```
- Request parsing tests with various edge cases (empty body, multiple headers, spaces)
- Response building tests with different status codes
- Connection handling tests with mock connections

### Router Tests
```bash
go test ./router/...
```
- Route matching tests for different patterns and prefixes
- 404 handling for unregistered routes
- Middleware chain execution tests

### Integration Tests
```bash
go test ./...
```
- Graceful shutdown tests with active connections
- End-to-end server functionality tests
- TLS connection tests

### Manual Testing
```bash
# Start server
go run main.go -directory ./test-files

# Test HTTP endpoints
curl -v http://localhost:4221/echo/test
curl -X POST -d "Hello World" http://localhost:4221/files/test.txt
curl http://localhost:4221/files/test.txt
curl http://localhost:4221/health

# Test HTTPS endpoints
curl -k https://localhost:4222/health
curl -k -H "Accept-Encoding: gzip" https://localhost:4222/echo/compressed
```

## 📊 Project Structure

```
├── main.go                    # Main entry point with flag parsing
├── config/
│   └── config.go             # Configuration and TLS certificate management
├── server/
│   └── server.go             # Main server logic with worker pool
├── http/
│   ├── request.go            # HTTP request parsing
│   ├── response.go           # HTTP response building
│   ├── request_test.go       # Request parser tests
│   └── response_test.go      # Response builder tests
├── router/
│   ├── router.go             # Routing logic with middleware support
│   └── router_test.go        # Router tests
├── handler/
│   ├── handler.go            # Handler interface and function adapter
│   ├── echo_handler.go       # Handler for /echo/*
│   ├── file_handler.go       # Handler for /files/*
│   ├── user_agent_handler.go # Handler for /user-agent
│   ├── health_handler.go     # Handler for /health with metrics
│   └── server_metrics.go     # Metrics interface definition
├── middleware/
│   └── middleware.go         # Middleware system with gzip and logging
└── main_test.go              # Integration tests
```

## 🎯 Learning Objectives

This project provides hands-on training with:
- **Low-level networking concepts** with TCP connections and TLS
- **HTTP protocol** deep understanding and request/response parsing
- **Go concurrency** with goroutines, channels, and worker pools
- **Modular architecture** with clean separation of concerns
- **Middleware patterns** for cross-cutting functionality
- **Unit and integration testing** with comprehensive coverage
- **Robust error handling** and graceful degradation
- **Security considerations** with input validation and TLS
- **Performance optimization** with connection pooling and compression

## 🔮 Future Enhancements

- [ ] HTTP/2 protocol support
- [ ] Request middleware system expansion (authentication, rate limiting)
- [ ] Configuration file support (YAML/JSON)
- [ ] Advanced logging and metrics (Prometheus integration)
- [ ] Request rate limiting and throttling
- [ ] Static file serving with caching headers
- [ ] WebSocket support for real-time communication
- [ ] Database integration examples
- [ ] Docker containerization
- [ ] Load balancing capabilities
- [ ] Built-in HTML templating engine
- [ ] Advanced security features (CORS, CSP headers)

## 📄 License

This project is part of a coding challenge and is intended for educational purposes.

---

**Note:** This server includes TLS support with self-signed certificates for development. For production use, replace with proper certificates and additional security hardening.