package router

import (
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
	"github.com/codecrafters-io/http-server-starter-go/middleware"
)

type Router struct {
	routes      map[string]handler.Handler
	middlewares []middleware.Middleware
}

func NewRouter() *Router {
	return &Router{
		routes: make(map[string]handler.Handler),
	}
}

func (r *Router) Handle(pattern string, handler handler.Handler) {
	r.routes[pattern] = handler
}

func (r *Router) ServeHTTP(request *http.Request, response *http.Response) {
	if request.Path == "/" {
		response.StatusCode = 200
		return
	}

	pathFound := false
	for pattern, handler := range r.routes {
		if strings.HasPrefix(request.Path, pattern+"/") || request.Path == pattern {
			handler.Handle(request, response)
			pathFound = true
			break
		}
	}

	if !pathFound {
		response.StatusCode = 404
	}

	for _, middleware := range r.middlewares {
		middleware(request, response)
	}
}

func (r *Router) Use(middlewares ...middleware.Middleware) {
	r.middlewares = middlewares
}
