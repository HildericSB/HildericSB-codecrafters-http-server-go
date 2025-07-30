package handler

import (
	"net/http"
	"strconv"

	httpPkg "github.com/codecrafters-io/http-server-starter-go/http"
)

type UserAgentHandler struct{}

func NewUserAgentHandler() *UserAgentHandler {
	return &UserAgentHandler{}
}

func (uah *UserAgentHandler) Handle(req *httpPkg.Request, res *httpPkg.Response) {
	userAgent := req.Headers["User-Agent"]
	if userAgent == "" {
		res.StatusCode = http.StatusBadRequest
		return
	}

	res.StatusCode = http.StatusOK
	res.Headers["Content-Type"] = "text/plain"
	res.Headers["Content-Length"] = strconv.Itoa(len(userAgent))
	res.Body = userAgent
}
