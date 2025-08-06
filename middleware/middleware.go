package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/handler"
	"github.com/codecrafters-io/http-server-starter-go/http"
)

type Middleware func(next handler.Handler) handler.Handler

// Chain represents a chain of middlewares
type Chain struct {
	middlewares []Middleware
}

func NewChain(middlewares []Middleware) *Chain {
	return &Chain{
		middlewares: middlewares,
	}
}

func (c *Chain) ContructMainHandler(h handler.Handler) handler.Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i](h)
	}
	return h
}

func GzipMiddleware() Middleware {
	return func(next handler.Handler) handler.Handler {
		return handler.HandlerFunc(func(req *http.Request, resp *http.Response) {
			next.Handle(req, resp)

			encodings := req.Headers["Accept-Encoding"]

			if strings.Contains(encodings, "gzip") {
				req.Headers["Content-Encoding"] = "gzip"

				var buf bytes.Buffer
				zw := gzip.NewWriter(&buf)

				_, err := zw.Write([]byte(resp.Body))
				if err != nil {
					fmt.Println("Error encoding the body", err)
					return
				}

				err = zw.Close()
				if err != nil {
					fmt.Println("Error closing the gzip writer", err)
					return
				}

				resp.Body = buf.String()
				resp.Headers["Content-Encoding"] = "gzip"
				resp.Headers["Content-Length"] = strconv.Itoa(len(resp.Body))
			}
		})
	}
}

func LoggingMiddleware() Middleware {
	return func(next handler.Handler) handler.Handler {
		return handler.HandlerFunc(func(req *http.Request, resp *http.Response) {
			startTime := time.Now()
			next.Handle(req, resp)
			elapsed := time.Since(startTime)

			log := fmt.Sprintf("[%s] %s %s - %d - %dms",
				startTime.Format("2006/01/02 15:04:05"),
				req.Method,
				req.Path,
				resp.StatusCode,
				elapsed.Milliseconds())

			fmt.Println(log)
		})
	}
}
