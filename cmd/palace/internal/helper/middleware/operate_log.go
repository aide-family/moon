package middleware

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport/http"

	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
)

type OperateLogParams struct {
	Operation     string
	Request       any
	Reply         any
	Error         error
	OriginRequest *http.Request
	Duration      time.Duration
}

type OperateLogFunc func(ctx context.Context, params *OperateLogParams)

func OperateLog(operateLogFunc OperateLogFunc) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			startTime := time.Now()
			reply, err := handler(ctx, req)
			duration := time.Since(startTime)
			var originRequest *http.Request
			if request, ok := http.RequestFromServerContext(ctx); ok {
				originRequest = request
			}
			params := &OperateLogParams{
				Operation:     permission.GetOperationByContextWithDefault(ctx),
				Request:       req,
				Reply:         reply,
				Error:         err,
				OriginRequest: originRequest,
				Duration:      duration,
			}
			operateLogFunc(ctx, params)
			return reply, err
		}
	}
}
