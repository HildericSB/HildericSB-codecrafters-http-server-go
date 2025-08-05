package handler

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/codecrafters-io/http-server-starter-go/http"
)

type HealthHandler struct {
	metrics ServerMetrics
}

type HealthResponse struct {
	Status            string `json:"status"`
	Timestamp         string `json:"timestamp"`
	Uptime            string `json:"uptime"`
	ActiveConnections int    `json:"active_connections"`
	TotalRequests     int    `json:"total_requests"`
}

func NewHealthHandler(metrics ServerMetrics) *HealthHandler {
	return &HealthHandler{
		metrics: metrics,
	}
}

func (eh *HealthHandler) Handle(req *http.Request, resp *http.Response) {
	resp.StatusCode = 200

	healthData := HealthResponse{
		Status:            "healthy",
		Timestamp:         time.Now().Format(time.RFC3339),
		Uptime:            time.Since(eh.metrics.ServerStartTime()).String(),
		ActiveConnections: eh.metrics.GetOpenConnections(),
		TotalRequests:     eh.metrics.GetTotalRequests(),
	}

	jsonBody, err := json.Marshal(healthData)
	if err != nil {
		fmt.Print("Error marshalling to json ", err)
	}

	resp.Body = string(jsonBody)
	resp.Headers["Content-Type"] = "application/json"
	resp.Headers["Content-Length"] = strconv.Itoa(len(resp.Body))
}
