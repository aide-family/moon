package middleware

import (
	"context"

	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
)

// SourceType 获取请求头中的Source-Type  sourceType System Team
func SourceType() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				sourceCode := tr.RequestHeader().Get("Source-Type")
				ctx = context.WithValue(ctx, sourceTypeKey{}, vobj.GetSourceType(sourceCode))
			}
			return handler(ctx, req)
		}
	}
}

// SourceTypeInfo Request header source
type sourceTypeKey struct{}

func GetSourceType(ctx context.Context) vobj.SourceType {
	sourceTypeInfo, ok := ctx.Value(sourceTypeKey{}).(vobj.SourceType)
	if ok {
		return sourceTypeInfo
	}
	return vobj.SourceTypeTeam
}
