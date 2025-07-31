package middleware

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"strconv"
	"strings"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

type Middleware func(req *http.Request, resp *http.Response)

func GzipMiddleware() Middleware {
	return func(req *http.Request, resp *http.Response) {
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
	}
}
