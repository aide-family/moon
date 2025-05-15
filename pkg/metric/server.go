package metric

import (
	"context"
	"strconv"
	"time"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	"github.com/go-kratos/kratos/v2/transport/http/status"
	"google.golang.org/grpc/codes"

	"github.com/aide-family/moon/pkg/util/timex"
)

func Server(name string, opts ...MetricServerOption) middleware.Middleware {
	server := &metricServer{
		name: name,
		handlers: []Handler{
			defaultServerHandler,
		},
	}
	for _, opt := range opts {
		opt(server)
	}
	serverHandlers := server.handlers
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int
				reason    string
				kind      string
				operation string
			)

			// default code
			code = status.FromGRPCCode(codes.OK)
			startTime := timex.Now()
			reply, err = handler(ctx, req)
			if len(serverHandlers) == 0 {
				return
			}
			if se := errors.FromError(err); se != nil {
				code = int(se.Code)
				reason = se.Reason
			}

			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}

			params := &request{
				code:      code,
				reason:    reason,
				kind:      kind,
				operation: operation,
				latency:   time.Since(startTime),
				server:    name,
			}
			for _, h := range serverHandlers {
				h(ctx, params)
			}
			return
		}
	}
}

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

type MetricServerOption func(*metricServer)

func WithServerHandler(handler Handler) MetricServerOption {
	return func(server *metricServer) {
		server.handlers = append(server.handlers, handler)
	}
}

func defaultServerHandler(_ context.Context, req Request) {
	labels := []string{
		req.GetKind(),
		req.GetOperation(),
		strconv.Itoa(int(req.GetCode())),
		req.GetReason(),
		req.GetServer(),
	}
	RequestTotalMetric.WithLabelValues(labels...).Inc()
	RequestLatencyMetric.WithLabelValues(labels...).Observe(float64(req.GetLatency().Milliseconds()))
}
