package middleware

import (
	"context"
	"strings"

	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

var sourceOperationList = []string{}

// SourceType 获取请求头中的Source-Type  sourceType System Team
func SourceType() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				sourceCode := tr.RequestHeader().Get("Source-Type")
				ctx = context.WithValue(ctx, sourceTypeKey{}, vobj.GetSourceType(sourceCode))
			}

			operation, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, merr.ErrorNotification("get operation failed")
			}
			// 如果是系统请求，且operation在列表中，则校验是否为系统管理员， 否则跳过校验
			if !GetSourceType(ctx).IsSystem() || GetTeamRole(ctx).IsAdminOrSuperAdmin() {
				return handler(ctx, req)
			}
			for _, op := range sourceOperationList {
				if strings.EqualFold(op, operation.Operation()) {
					return nil, merr.ErrorI18nForbidden(ctx)
				}
			}

			return handler(ctx, req)
		}
	}
}

// SourceTypeInfo Request header source
type sourceTypeKey struct{}

// GetSourceType get source type
func GetSourceType(ctx context.Context) vobj.SourceType {
	sourceTypeInfo, ok := ctx.Value(sourceTypeKey{}).(vobj.SourceType)
	if ok {
		return sourceTypeInfo
	}
	return vobj.SourceTypeTeam
}
