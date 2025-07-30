package handler

import (
	"net/http"
	"strconv"
	"strings"

	httpPkg "github.com/codecrafters-io/http-server-starter-go/http"
)

type EchoHandler struct{}

func NewEchoHandler() *EchoHandler {
	return &EchoHandler{}
}

func (eh *EchoHandler) Handle(req *httpPkg.Request, res *httpPkg.Response) {
	body := strings.TrimPrefix(req.Path, "/echo/")
	res.StatusCode = http.StatusOK
	res.Headers["Content-Type"] = "text/plain"
	res.Headers["Content-Length"] = strconv.Itoa(len(body))
	res.Body = body
}
