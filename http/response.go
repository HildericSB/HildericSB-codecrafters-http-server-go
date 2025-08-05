package http

import (
	"fmt"
	"net"
	"strconv"

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

	if request.Headers["Connection"] == "close" {
		response.Headers["Connection"] = "close"
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
		return fmt.Errorf("unknown status code: %d", r.StatusCode)
	}

	body := r.Body

	rep := "HTTP/1.1 " + strconv.Itoa(r.StatusCode) + " " + statusMessage + config.CRLF
	for k, v := range r.Headers {
		rep = rep + k + ":" + v + config.CRLF
	}

	rep = rep + config.CRLF + body

	_, err := r.Connection.Write([]byte(rep))
	if err != nil {
		fmt.Println("Error writing http response to client ", err)
		return err
	}

	return nil
}
