package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

type HealthHandler struct {
	metrics ServerMetrics
}

func NewHealthHandler(metrics ServerMetrics) *HealthHandler {
	return &HealthHandler{
		metrics: metrics,
	}
}

func (eh *HealthHandler) Handle(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200
	resp.Body = fmt.Sprintf(`
	{
		"status": "healthy",
		"timestamp": "%v",
		"uptime": "%v",
		"active connections": "%v",
		"total_requests": "%v",
	}`,
		time.Now().Format(time.RFC1123),
		time.Since(eh.metrics.ServerStartTime()),
		eh.metrics.GetOpenConnections(),
		eh.metrics.GetTotalRequests(),
	)

	resp.Headers["Content-Type"] = "application/json"
	resp.Headers["Content-Length"] = strconv.Itoa(len(resp.Body))
}
