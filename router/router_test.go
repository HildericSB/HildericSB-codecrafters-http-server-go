package router

import (
	"testing"

	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
)

func TestRouterServeHTTP_RootPath(t *testing.T) {
	r := NewRouter()
	request := &http.Request{Path: "/", Headers: map[string]string{}}
	response := &http.Response{Headers: map[string]string{}}

	r.ServeHTTP(request, response)

	if response.StatusCode != 200 {
		t.Errorf("Expected status 200 for root path, got %d", response.StatusCode)
	}
}

func TestRouterServeHTTP_RegisteredRoute(t *testing.T) {
	r := NewRouter()
	handlerCalled := false
	r.Handle("/echo", handler.HandlerFunc(func(req *http.Request, res *http.Response) {
		handlerCalled = true
		res.StatusCode = 201
	}))
	request := &http.Request{Path: "/echo/hello", Headers: map[string]string{}}
	response := &http.Response{Headers: map[string]string{}}

	r.ServeHTTP(request, response)

	if !handlerCalled {
		t.Error("Handler was not called for registered route")
	}
	if response.StatusCode != 201 {
		t.Errorf("Expected status 201 from handler, got %d", response.StatusCode)
	}
}

func TestRouterServeHTTP_UnregisteredRoute(t *testing.T) {
	r := NewRouter()
	request := &http.Request{Path: "/notfound", Headers: map[string]string{}}
	response := &http.Response{Headers: map[string]string{}}

	r.ServeHTTP(request, response)

	if response.StatusCode != 404 {
		t.Errorf("Expected status 404 for unregistered route, got %d", response.StatusCode)
	}
}
