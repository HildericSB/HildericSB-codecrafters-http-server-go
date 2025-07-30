// TestGracefulShutdown_WithActiveConnections tests shutdown with active connections

package main

import (
	"net"
	"testing"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/server"
)

func TestGracefulShutdown_WithActiveConnections(t *testing.T) {
	server, err := server.NewServerWithDefaults()
	if err != nil {
		t.Fatalf("Failed to create server: %v", err)
	}

	go server.Start()
	time.Sleep(100 * time.Millisecond) //  time for server to start

	port := server.Port

	numberOfConnections := 5
	connections := make([]net.Conn, numberOfConnections)
	for i := range numberOfConnections {
		conn, err := net.Dial("tcp", ":"+port)
		if err != nil {
			t.Fatalf("Error creating TCP connection \n")
		}
		connections[i] = conn
	}

	time.Sleep(100 * time.Millisecond) //  time for connections to established

	activeConnections := server.GetOpenConnections()
	if activeConnections != numberOfConnections {
		t.Fatalf("Error number of active connections, activeConnections : %d\n", activeConnections)
	}

	// Simulate connections finishing gracefully
	for _, connection := range connections {
		connection.Write([]byte("GET / HTTP/1.1\r\nHost: localhost\r\nConnection: close\r\n\r\n"))
		connection.Close()
	}

	start := time.Now()
	server.ShutDown()
	duration := time.Since(start)

	if duration > time.Second {
		t.Errorf("Shutdown took too long: %v", duration)
	}
}
