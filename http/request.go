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
	_, err := conn.Read(buffer)
	if err != nil {
		if err != io.EOF {
			fmt.Println("Error reading client request from connection")
			return nil, err
		}
		// No data from client, so no request, no error
		return nil, nil
	}

	req := string(buffer)
	lines := strings.Split(req, "\r\n")

	//Read path and method type
	request := Request{
		Path:       strings.Split(lines[0], " ")[1],
		Method:     strings.Split(lines[0], " ")[0],
		Connection: conn,
	}

	// Read headers and body
	headers := make(map[string]string)

	for i, line := range lines[1:] {
		// If line is empty, there is no more header, next line is the body
		if line == "" {
			if len(lines) > i+2 {
				request.Body = lines[i+2]
			}
			break
		}

		headerSplit := strings.SplitN(line, ":", 2)
		for i, v := range headerSplit {
			headerSplit[i] = strings.TrimSpace(v)
		}
		if len(headerSplit) == 2 {
			headers[headerSplit[0]] = headerSplit[1]
		}
	}

	request.Headers = headers

	return &request, nil
}
