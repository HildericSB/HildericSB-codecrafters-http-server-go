package handler

import (
	"github.com/codecrafters-io/http-server-starter-go/http"
)

type Handler interface {
	Handle(req *http.Request, resp *http.Response)
}

type handlerfunc struct {
	fn func(request *http.Request, response *http.Response)
}

func (h handlerfunc) Handle(request *http.Request, response *http.Response) {
	h.fn(request, response)
}

func HandlerFunc(handlerFn func(request *http.Request, response *http.Response)) Handler {
	return handlerfunc{fn: handlerFn}
}
