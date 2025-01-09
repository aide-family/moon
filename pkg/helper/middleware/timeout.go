package middleware

import (
	"context"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
)

// Timeout ctx timeout
func Timeout(t time.Duration) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if t <= 0 {
				return handler(ctx, req)
			}
			ctx, cancel := context.WithTimeout(ctx, t)
			defer cancel()
			return handler(ctx, req)
		}
	}
}
