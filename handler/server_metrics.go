package handler

import "time"

type ServerMetrics interface {
	ServerStartTime() time.Time
	GetOpenConnections() int
	GetTotalRequests() int
}
