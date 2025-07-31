package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

type HealthHandler struct {
	serverStartTime *time.Time
}

func NewHealthHandler(ssT *time.Time) *HealthHandler {
	return &HealthHandler{
		serverStartTime: ssT,
	}
}

func (eh *HealthHandler) Handle(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
	resp.Body = fmt.Sprintf(`
	{
		"status": "healthy",
		"timestamp": "%v",
		"uptime": "%v",
	}
	`, time.Now().Format(time.RFC1123), time.Since(*eh.serverStartTime))

	resp.Headers["Content-Type"] = "application/json"
	resp.Headers["Content-Length"] = strconv.Itoa(len(resp.Body))
}
