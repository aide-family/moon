package middler

import (
	"context"
	"strconv"

	"github.com/aide-family/moon/pkg/middler/permission"
	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/validate"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// ExtractMetadata extracts the teamID and token from the metadata or HTTP header and sets them to the context.
func ExtractMetadata() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// First try to get the teamID from the metadata.
			if md, ok := metadata.FromServerContext(ctx); ok {
				if teamID := md.Get(cnst.MetadataGlobalKeyTeamID); validate.TextIsNotNull(teamID) {
					if teamID, err := strconv.ParseUint(teamID, 10, 32); err == nil {
						ctx = permission.WithTeamIDContext(ctx, uint32(teamID))
					}
				}
				if token := md.Get(cnst.MetadataGlobalKeyToken); validate.TextIsNotNull(token) {
					ctx = permission.WithTokenContext(ctx, token)
				}
			}

			// If the metadata does not have the teamID, try to get it from the HTTP header.
			if tr, ok := transport.FromServerContext(ctx); ok {
				if xTeamID := tr.RequestHeader().Get(cnst.XHeaderTeamID); xTeamID != "" {
					if teamID, err := strconv.ParseUint(xTeamID, 10, 32); validate.IsNil(err) {
						ctx = permission.WithTeamIDContext(ctx, uint32(teamID))
					}
				}
				if xToken := tr.RequestHeader().Get(cnst.XHeaderToken); xToken != "" {
					ctx = permission.WithTokenContext(ctx, xToken)
				}
			}

			return handler(ctx, req)
		}
	}
}
