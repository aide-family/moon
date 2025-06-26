package metric

import (
	"context"
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
