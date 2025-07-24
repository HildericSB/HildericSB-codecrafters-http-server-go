package http

import (
	"fmt"
	"io"
	"net"
	"testing"
)

// testConn is a mock connection that implements io.Reader for testing.
type testConn struct {
	net.Conn
	data string
	pos  int
}

func (c *testConn) Read(p []byte) (int, error) {
	if c.pos >= len(c.data) {
		return 0, io.EOF
	}
	n := copy(p, c.data[c.pos:])
	c.pos += n
	return n, nil
}

func TestParseRequest_EmptyBody(t *testing.T) {
	data := "GET /empty HTTP/1.1\r\nHost: localhost\r\n\r\n"
	conn := &testConn{data: data}
	req, err := ParseRequest(conn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if req.Method != "GET" || req.Path != "/empty" {
		t.Errorf("Expected GET /empty, got %s %s", req.Method, req.Path)
	}
	if req.Body != "" {
		t.Errorf("Expected empty body, got '%s'", req.Body)
	}
}

func TestParseRequest_MultipleHeaders(t *testing.T) {
	data := "POST /api HTTP/1.1\r\nHost: example.com\r\nContent-Type: application/json\r\nAuthorization: Bearer token123\r\n\r\n"
	conn := &testConn{data: data}
	req, err := ParseRequest(conn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if req.Headers["Host"] != "example.com" {
		t.Errorf("Expected Host 'example.com', got '%s'", req.Headers["Host"])
	}
	if req.Headers["Content-Type"] != "application/json" {
		t.Errorf("Expected Content-Type 'application/json', got '%s'", req.Headers["Content-Type"])
	}
	if req.Headers["Authorization"] != "Bearer token123" {
		t.Errorf("Expected Authorization 'Bearer token123', got '%s'", req.Headers["Authorization"])
	}
}

func TestParseRequest_LongBody(t *testing.T) {
	body := "This is a longer body with multiple words and sentences."
	data := fmt.Sprintf("PUT /upload HTTP/1.1\r\nContent-Length: %d\r\n\r\n%s", len(body), body)
	conn := &testConn{data: data}
	req, err := ParseRequest(conn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if req.Body != body {
		t.Errorf("Expected body '%s', got '%s'", body, req.Body)
	}
}

func TestParseRequest_HeadersWithSpaces(t *testing.T) {
	data := "GET /test HTTP/1.1\r\nUser-Agent:   Mozilla/5.0   \r\nAccept:  text/html  \r\n\r\n"
	conn := &testConn{data: data}
	req, err := ParseRequest(conn)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if req.Headers["User-Agent"] != "Mozilla/5.0" {
		t.Errorf("Expected User-Agent 'Mozilla/5.0', got '%s'", req.Headers["User-Agent"])
	}
	if req.Headers["Accept"] != "text/html" {
		t.Errorf("Expected Accept 'text/html', got '%s'", req.Headers["Accept"])
	}
}

func TestParseRequest(t *testing.T) {
	// Use net.Pipe for in-memory connection
	client, server := net.Pipe()
	defer client.Close()
	defer server.Close()

	// Simulate a simple HTTP GET request
	go func() {
		req := "GET /abcd HTTP/1.1\r\nHost: localhost\r\nUser-Agent: test-agent\r\n\r\n"
		client.Write([]byte(req))
	}()

	request, err := ParseRequest(server)
	if err != nil {
		t.Fatalf("failed to parse request: %v", err)
	}

	if request == nil {
		t.Fatalf("expected request, got nil")
	}

	if request.Path != "/abcd" {
		t.Errorf("expected Path '/abcd', got '%s'", request.Path)
	}
	if request.Method != "GET" {
		t.Errorf("expected Method 'GET', got '%s'", request.Method)
	}
	if ua := request.Headers["User-Agent"]; ua != "test-agent" {
		t.Errorf("expected User-Agent 'test-agent', got '%s'", ua)
	}
}
