package middleware

import (
	"context"
	"strconv"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/validate"
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
	if xTeamID := tr.RequestHeader().Get(cnst.XHeaderTeamID); xTeamID != "" {
		if teamID, err := strconv.ParseUint(xTeamID, 10, 32); validate.IsNil(err) {
			ctx = permission.WithTeamIDContext(ctx, uint32(teamID))
		}
	}
	return ctx, nil
}
