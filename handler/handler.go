package handler

import (
	"github.com/codecrafters-io/http-server-starter-go/http"
)

type Handler interface {
	Handle(req *http.Request, resp *http.Response)
}

// Using Method on function type to allow both struct based handlers and
// function based handlers
type HandlerFunc func(req *http.Request, resp *http.Response)

func (h HandlerFunc) Handle(req *http.Request, resp *http.Response) {
	h(req, resp)
}
