package http

import (
	"net"
	"testing"
)

type dummyConn struct{ net.Conn }

func (d *dummyConn) Write(b []byte) (int, error) { return len(b), nil }

func TestNewResponse_ConnectionCloseHeader(t *testing.T) {
	req := &Request{
		Headers:    map[string]string{"Connection": "close"},
		Connection: &dummyConn{},
	}
	resp := NewResponse(req)
	if resp.Headers["Connection"] != "close" {
		t.Error("Expected Connection header to be 'close'")
	}
}

func TestNewResponse_ConnectionAssignment(t *testing.T) {
	dc := &dummyConn{}
	req := &Request{Headers: map[string]string{}, Connection: dc}
	resp := NewResponse(req)
	if resp.Connection != dc {
		t.Error("Expected response.Connection to match request.Connection")
	}
}
