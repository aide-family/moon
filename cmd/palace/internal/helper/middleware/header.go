package middleware

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/cnst"
)

func BindHeaders() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			ctx, err := withAllHeaders(ctx)
			if err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}
	}
}

func withAllHeaders(ctx context.Context) (context.Context, error) {
	tr, ok := transport.FromServerContext(ctx)
	if !ok {
		return nil, merr.ErrorBadRequest("not allow request")
	}

	ctx = permission.WithOperationContext(ctx, tr.Operation())
	if teamIDStr := tr.RequestHeader().Get(cnst.XHeaderTeamID); teamIDStr != "" {
		teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
		if err == nil {
			ctx = permission.WithTeamIDContext(ctx, uint32(teamID))
		}
	}
	return ctx, nil
}
