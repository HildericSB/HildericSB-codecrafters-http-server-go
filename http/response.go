package http

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/config"
)

type Response struct {
	StatusCode int
	Body       string
	Headers    map[string]string
	Connection net.Conn
}

func NewResponse(request *Request) *Response {
	response := &Response{
		Headers: map[string]string{},
	}

	pathSplit := strings.Split(request.Path, "/")
	pathSplitLength := len(pathSplit)

	if request.Path == "/" {
		response.StatusCode = 200
	} else if pathSplitLength >= 2 {
		switch pathSplit[1] {
		case "echo":
			HandleEcho(request, response)
		case "user-agent":
			HandleUserAgent(request, response)
		case "files":
			if request.Method == "GET" {
				HandleFileRead(request, response)
			}

			if request.Method == "POST" {
				HandleFileUpload(request, response)
			}
		}

	}

	if request.Headers["Connection"] == "close" {
		response.Headers["Connection"] = "close"
	}

	if response.StatusCode == 0 {
		response.StatusCode = 404
	}

	response.Connection = request.Connection

	return response
}

func (r *Response) SendToClient(request *Request) error {
	var statusMessage string
	switch r.StatusCode {
	case 404:
		statusMessage = "Not Found"
	case 201:
		statusMessage = "Created"
	case 400:
		statusMessage = "Bad request"
	case 200:
		statusMessage = "OK"
	default:
		panic("HTTP statusCode unknown")
	}

	body := r.Body
	encodings := request.Headers["Accept-Encoding"]

	if strings.Contains(encodings, "gzip") {
		r.Headers["Content-Encoding"] = "gzip"
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		_, err := zw.Write([]byte(body))
		if err != nil {
			fmt.Println("Error encoding the body")
		}
		err = zw.Close()
		if err != nil {
			return err
		}

		body = buf.String()
		r.Headers["Content-Length"] = strconv.Itoa(len(body))
	}

	rep := "HTTP/1.1 " + strconv.Itoa(r.StatusCode) + " " + statusMessage + config.CRLF
	for k, v := range r.Headers {
		rep = rep + k + ":" + v + config.CRLF
	}

	rep = rep + config.CRLF + body

	fmt.Println("\n HTTP reponse : \n" + rep)

	_, err := r.Connection.Write([]byte(rep))
	if err != nil {
		fmt.Println("Error writing http response to client ", err)
		return err
	}

	return nil
}
