package handler

import (
	"github.com/codecrafters-io/http-server-starter-go/http"
)

type Handler interface {
	Handle(req *http.Request, resp *http.Response)
}

type Handlerfunc struct {
	Fn func(request *http.Request, response *http.Response)
}

func (h Handlerfunc) Handle(request *http.Request, response *http.Response) {
	h.Fn(request, response)
}

func HandlerFunc(handlerFn func(request *http.Request, response *http.Response)) Handler {
	return Handlerfunc{Fn: handlerFn}
}
