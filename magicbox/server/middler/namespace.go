package middler

import (
	"context"

	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
)

func MustNamespace() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)
				namespaceUID, err := snowflake.ParseString(namespace)
				if err != nil {
					return nil, merr.ErrorForbidden("namespace is invalid, please set the namespace in the request header, Example: %s: 1", cnst.HTTPHeaderXNamespace)
				}
				ctx = contextx.WithNamespace(ctx, namespaceUID)
				tr.RequestHeader().Set(cnst.MetadataGlobalKeyNamespace, namespace)

				if namespaceUID > 0 {
					return handler(ctx, req)
				}
			}

			if md, ok := metadata.FromServerContext(ctx); ok {
				namespace := md.Get(cnst.MetadataGlobalKeyNamespace)
				namespaceUID, err := snowflake.ParseString(namespace)
				if err != nil {
					return nil, merr.ErrorForbidden("namespace is invalid, please set the namespace in the metadata, Example: %s: 1", cnst.HTTPHeaderXNamespace)
				}
				ctx = contextx.WithNamespace(ctx, namespaceUID)
				md.Set(cnst.MetadataGlobalKeyNamespace, namespace)
				if namespaceUID > 0 {
					return handler(ctx, req)
				}
			}

			return nil, merr.ErrorForbidden("namespace is required, please set the namespace in the request header or metadata, Example: %s: 1", cnst.HTTPHeaderXNamespace)
		}
	}
}

// MustNamespaceExist 检查namespace必须存在且有效
func MustNamespaceExist(hasNamespace func(ctx context.Context) (snowflake.ID, error)) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			namespace, err := hasNamespace(ctx)
			if err != nil {
				return nil, err
			}
			ctx = contextx.WithNamespace(ctx, namespace)
			return handler(ctx, req)
		}
	}
}
