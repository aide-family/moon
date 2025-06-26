package metric

import (
	"context"
	"time"
)

type Handler func(ctx context.Context, req Request)

type metricServer struct {
	handlers []Handler
	name     string
}

type Request interface {
	GetServer() string
	GetCode() int
	GetReason() string
	GetKind() string
	GetOperation() string
	GetLatency() time.Duration
}

var _ Request = (*request)(nil)

type request struct {
	server    string
	code      int
	reason    string
	kind      string
	operation string
	latency   time.Duration
}

// GetServer implements Request.
func (r *request) GetServer() string {
	return r.server
}

// GetCode implements Request.
func (r *request) GetCode() int {
	return r.code
}

// GetKind implements Request.
func (r *request) GetKind() string {
	return r.kind
}

// GetLatency implements Request.
func (r *request) GetLatency() time.Duration {
	return r.latency
}

// GetOperation implements Request.
func (r *request) GetOperation() string {
	return r.operation
}

// GetReason implements Request.
func (r *request) GetReason() string {
	return r.reason
}
