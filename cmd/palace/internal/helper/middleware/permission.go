package middleware

import (
	"context"

	"github.com/go-kratos/kratos/v2/middleware"
)

type permissionValidateFunc func(ctx context.Context) error

// MustPermission must permission validate
func MustPermission(validate permissionValidateFunc) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if err := validate(ctx); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}
	}
}
