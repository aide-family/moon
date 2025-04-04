package middleware

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/pkg/helper/metric"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// Logging is an server logging middleware.
func Logging(logger log.Logger) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			var (
				code      int32
				reason    string
				kind      string
				operation string
			)
			startTime := time.Now()
			if info, ok := transport.FromServerContext(ctx); ok {
				kind = info.Kind().String()
				operation = info.Operation()
			}
			metric.IncRequestCounter(kind, operation)
			reply, err = handler(ctx, req)
			if err != nil {
				code = 500
				reason = "未知系统错误"
				if se := errors.FromError(err); se != nil {
					code = se.Code
					reason = se.Reason
				}
				metric.IncCounterRequestErr(kind, operation, code)
			}
			latency := time.Since(startTime)
			metric.RecordResponseTime(kind, operation, latency.Seconds())

			level, stack := extractError(err)
			_ = log.WithContext(ctx, logger).Log(level,
				"kind", "server",
				"component", kind,
				"operation", operation,
				"args", extractArgs(req),
				"code", code,
				"reason", reason,
				"stack", stack,
				"latency", latency.String(),
			)
			return
		}
	}
}

// extractError returns the string of the error
func extractError(err error) (log.Level, string) {
	if err != nil {
		return log.LevelError, err.Error()
	}
	return log.LevelInfo, ""
}

// extractArgs returns the string of the req
func extractArgs(req any) string {
	if stringer, ok := req.(fmt.Stringer); ok {
		return stringer.String()
	}
	bytes, err := types.Marshal(req)
	if err != nil {
		return fmt.Sprintf("%+v", req)
	}
	return string(bytes)
}
