package http

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/config"
)

type Request struct {
	Path       string
	Headers    map[string]string
	Connection net.Conn
	Method     string
	Body       string
}

func ParseRequest(conn net.Conn) (*Request, error) {
	// Create a buffer and read the HTTP request from connection
	buffer := make([]byte, config.BUFFER_SIZE)
	fmt.Println("conn : ", conn)
	n, err := conn.Read(buffer)
	if err != nil {
		if err == io.EOF {
			return nil, err // No more data client closed the connection
		}
		return nil, fmt.Errorf("failed to read from connection: %w", err)
	}

	if n == 0 {
		return nil, nil // No data received
	}

	req := string(buffer[:n])
	lines := strings.Split(req, "\r\n")

	if len(lines) == 0 {
		return nil, fmt.Errorf("empty request")
	}

	//Read path and method type
	request := Request{
		Path:       strings.Split(lines[0], " ")[1],
		Method:     strings.Split(lines[0], " ")[0],
		Connection: conn,
		Headers:    make(map[string]string),
	}

	// Parse headers
	for _, line := range lines[1:] {
		// If line is empty, there is no more header, next line is the body
		if line == "" {
			break
		}

		headerSplit := strings.SplitN(line, ":", 2)
		for i, v := range headerSplit {
			headerSplit[i] = strings.TrimSpace(v)
		}
		if len(headerSplit) == 2 {
			request.Headers[headerSplit[0]] = headerSplit[1]
		}
	}

	// Read Body
	bodyStart := strings.Index(req, "\r\n\r\n") + 4
	if bodyStart < len(req) {
		request.Body = req[bodyStart:]
	}

	return &request, nil
}
