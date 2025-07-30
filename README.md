# Simple HTTP Server in Go

A lightweight HTTP server implementation built from scratch in Go as part of the [CodeCrafters](https://codecrafters.io) "Build Your Own HTTP Server" challenge. This project serves as Go training by creating a custom web server.

## 🚀 Features

- **HTTP/1.1 Protocol Support** - Handles basic HTTP requests and responses
- **Concurrent Connections** - Each connection handled in a separate goroutine
- **File Operations** - Upload and download files via HTTP endpoints
- **Gzip Compression** - Automatic response compression when supported by client
- **Echo Endpoint** - Simple endpoint for testing and debugging
- **User-Agent Detection** - Endpoint to retrieve client user-agent information
- **Persistent Connections** - Support for keep-alive connections
- **Command-Line Configuration** - Configurable file directory via flags
- **Graceful Shutdown** - Proper server shutdown handling with active connections

## 📋 Supported Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/` | Returns 200 OK (health check) |
| `GET` | `/echo/{message}` | Returns the message in the response body |
| `GET` | `/user-agent` | Returns the client's User-Agent header |
| `GET` | `/files/{filename}` | Downloads a file from the server |
| `POST` | `/files/{filename}` | Uploads a file to the server |

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

## 📡 API Examples

### Basic Health Check
```bash
curl http://localhost:4221/
# Response: 200 OK
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

## 🏗️ Architecture

### Core Components

#### `server` Package
- **Server Manager**: Manages the main TCP server and graceful shutdown
- **Connection Handler**: Tracks open connections for proper shutdown

#### `http` Package
- **Request Parser**: Parses incoming HTTP requests into structured data
- **Response Builder**: Constructs HTTP responses with proper headers
- **Compression Support**: Automatic gzip based on Accept-Encoding headers

#### `router` Package
- **HTTP Router**: Dispatches requests to appropriate handlers
- **Pattern Matching**: Support for prefix-based route matching

#### `handler` Package
- **Handler Interface**: Common interface for all request handlers
- **EchoHandler**: Handles `/echo/*` endpoints
- **FileHandler**: Handles file operations (`/files/*`)
- **UserAgentHandler**: Handles `/user-agent` endpoint

#### `config` Package
- **Configuration**: Server constants and default configuration

### Request Flow
1. TCP connection established
2. HTTP request parsed from raw bytes
3. Request routed to appropriate handler
4. Response generated with proper headers
5. Optional gzip compression applied
6. Response sent back to client

## 🔧 Technical Details

### HTTP Features Implemented
- ✅ HTTP/1.1 protocol parsing
- ✅ Request method handling (GET, POST)
- ✅ Header parsing and validation
- ✅ Request body handling
- ✅ Status code responses (200, 201, 400, 404, 500)
- ✅ Content-Type and Content-Length headers
- ✅ Connection management (keep-alive/close)
- ✅ Gzip compression support

### Concurrency
- Each client connection handled in a separate goroutine
- Thread-safe file operations
- Graceful connection cleanup with defer statements

### Error Handling
- Proper error responses for malformed requests
- File operation error handling
- Connection error recovery
- File name validation to prevent directory traversal

### Security
- Directory traversal protection in file handlers
- Content-Length header validation
- Safe file path handling

## 🧪 Testing

The project includes comprehensive unit tests for core components:

### HTTP Tests
```bash
go test ./http/...
```
- Request parsing tests with various cases (empty body, multiple headers, etc.)
- Response building tests

### Router Tests
```bash
go test ./router/...
```
- Route matching tests for different patterns
- 404 handling for unregistered routes

### Integration Tests
```bash
go test ./...
```
- Graceful shutdown tests with active connections

### Manual Testing
```bash
# Start server
go run main.go -directory ./test-files

# Test in another terminal
curl -v http://localhost:4221/echo/test
curl -X POST -d "Hello World" http://localhost:4221/files/test.txt
curl http://localhost:4221/files/test.txt
```

## 📊 Project Structure

```
├── main.go                    # Main entry point
├── config/
│   └── config.go             # Configuration and constants
├── server/
│   └── server.go             # Main server logic
├── http/
│   ├── request.go            # HTTP request parsing
│   ├── response.go           # HTTP response building
│   ├── request_test.go       # Request parser tests
│   └── response_test.go      # Response builder tests
├── router/
│   ├── router.go             # Routing logic
│   └── router_test.go        # Router tests
└── handler/
    ├── handler.go            # Handler interface
    ├── echo_handler.go       # Handler for /echo/*
    ├── file_handler.go       # Handler for /files/*
    └── user_agent_handler.go # Handler for /user-agent
```

## 🎯 Learning Objectives

This project provides hands-on training with:
- **Low-level networking concepts** with TCP connections
- **HTTP protocol** understanding and request/response parsing
- **Go concurrency** with goroutines
- **Modular architecture** with separation of concerns
- **Unit and integration testing**
- **Robust error handling**
- **Security considerations** with input validation

## 🔮 Future Enhancements

- [ ] HTTPS/TLS support
- [ ] HTTP/2 protocol support
- [ ] Request middleware system
- [ ] Configuration file support
- [ ] Logging and metrics
- [ ] Request rate limiting
- [ ] Static file serving with caching
- [ ] WebSocket support
- [ ] Authentication and authorization
- [ ] Built-in HTML templating

## 📄 License

This project is part of a coding challenge and is intended for educational purposes.

---

**Note:** This server is designed for learning purposes and should not be used in production without additional security and performance improvements.