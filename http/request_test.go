package http

import (
	"net"
	"testing"
)

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
