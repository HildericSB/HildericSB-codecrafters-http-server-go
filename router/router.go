package router

import (
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
	"github.com/codecrafters-io/http-server-starter-go/middleware"
)

type Router struct {
	routes map[string]handler.Handler
	chain  middleware.Chain
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
	mainHandler := handler.HandlerFunc(func(req *http.Request, res *http.Response) {
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
	})

	// TODO : avoid constructing the main handler everytime
	finalHandler := r.chain.ContructMainHandler(mainHandler)
	finalHandler.Handle(request, response)
}

func (r *Router) Use(middlewares ...middleware.Middleware) {
	r.chain = *middleware.NewChain(middlewares)
}
