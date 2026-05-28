package middler

import (
	"context"
	"strconv"
	"strings"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// AuthConfig configures exclusive JWT or service-key authentication per request.
type AuthConfig struct {
	AllowedServiceKeys []string
	JWTSecret          string
	JWTClaims          jwtv5.Claims
	ValidateUser       func(ctx context.Context, userUID snowflake.ID) error
}

func MustAuth(cfg AuthConfig) middleware.Middleware {
	jwtMiddlewares := []middleware.Middleware{
		JwtServe(cfg.JWTSecret, cfg.JWTClaims),
		MustLogin(),
		BindJwtToken(),
	}
	if cfg.ValidateUser != nil {
		jwtMiddlewares = append(jwtMiddlewares, ValidateUser(cfg.ValidateUser))
	}
	serviceKeyMiddlewares := []middleware.Middleware{
		ServiceKeyServe(cfg.AllowedServiceKeys),
		BindServiceKeyToken(),
	}

	return func(handler middleware.Handler) middleware.Handler {
		jwtHandler := chainMiddleware(jwtMiddlewares, handler)
		serviceKeyHandler := chainMiddleware(serviceKeyMiddlewares, handler)

		return func(ctx context.Context, req any) (any, error) {
			ensureRequestAuthorizationHeader(ctx)
			_, credential, ok := authorizationFromContext(ctx)
			if ok && IsServiceKeyCredential(credential) {
				return serviceKeyHandler(ctx, req)
			}
			if !ok {
				return nil, merr.ErrorUnauthorized("authorization is required")
			}
			ctx = contextx.WithAuthMode(ctx, contextx.AuthModeJWT)
			return jwtHandler(ctx, req)
		}
	}
}

// AuthClient propagates the single auth mode already established in context.
// Service-key chains forward only sk-xxx; JWT chains forward only Bearer JWT.
func AuthClient(headers ...string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			clientContext, ok := transport.FromClientContext(ctx)
			if !ok {
				return handler(ctx, req)
			}

			switch contextx.GetAuthMode(ctx) {
			case contextx.AuthModeServiceKey:
				propagateServiceKeyAuth(ctx, clientContext, headers...)
			case contextx.AuthModeJWT:
				propagateJWTAuth(ctx, clientContext, headers...)
			default:
				if _, credential, ok := authorizationFromContext(ctx); ok && IsServiceKeyCredential(credential) {
					propagateServiceKeyAuth(ctx, clientContext, headers...)
				} else {
					propagateJWTAuth(ctx, clientContext, headers...)
				}
			}

			return handler(ctx, req)
		}
	}
}

func propagateServiceKeyAuth(ctx context.Context, clientContext transport.Transporter, headers ...string) {
	if auth := resolveOutboundAuthorization(ctx, true); auth != "" {
		clientContext.RequestHeader().Set(cnst.HTTPHeaderAuthorization, auth)
	}
	propagateNamespace(ctx, clientContext, headers...)
}

func propagateJWTAuth(ctx context.Context, clientContext transport.Transporter, headers ...string) {
	if auth := resolveOutboundAuthorization(ctx, false); auth != "" {
		clientContext.RequestHeader().Set(cnst.HTTPHeaderAuthorization, auth)
	}
	propagateNamespace(ctx, clientContext, headers...)
}

func resolveOutboundAuthorization(ctx context.Context, serviceKeyOnly bool) string {
	full, credential, ok := authorizationFromContext(ctx)
	if !ok {
		return ""
	}
	if serviceKeyOnly && !IsServiceKeyCredential(credential) {
		return ""
	}
	if !serviceKeyOnly && IsServiceKeyCredential(credential) {
		return ""
	}
	return full
}

func propagateNamespace(ctx context.Context, clientContext transport.Transporter, headers ...string) {
	if namespace := resolveOutboundNamespace(ctx); namespace != "" {
		clientContext.RequestHeader().Set(cnst.HTTPHeaderXNamespace, namespace)
	}
	if tr, ok := transport.FromServerContext(ctx); ok {
		for _, header := range headers {
			if value := tr.RequestHeader().Get(header); value != "" {
				clientContext.RequestHeader().Set(header, value)
			}
		}
	}
	if md, ok := metadata.FromClientContext(ctx); ok {
		for _, header := range headers {
			if value := md.Get(header); value != "" {
				clientContext.RequestHeader().Set(header, value)
			}
		}
	}
}

func resolveOutboundNamespace(ctx context.Context) string {
	if tr, ok := transport.FromServerContext(ctx); ok {
		if namespace := strings.TrimSpace(tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace)); namespace != "" {
			return namespace
		}
		if namespace := strings.TrimSpace(tr.RequestHeader().Get(cnst.MetadataGlobalKeyNamespace)); namespace != "" {
			return namespace
		}
	}
	if md, ok := metadata.FromClientContext(ctx); ok {
		if namespace := strings.TrimSpace(md.Get(cnst.MetadataGlobalKeyNamespace)); namespace != "" {
			return namespace
		}
	}
	if md, ok := metadata.FromServerContext(ctx); ok {
		if namespace := strings.TrimSpace(md.Get(cnst.MetadataGlobalKeyNamespace)); namespace != "" {
			return namespace
		}
	}
	if namespaceUID, ok := contextx.TryGetNamespace(ctx); ok {
		return strconv.FormatInt(namespaceUID.Int64(), 10)
	}
	return ""
}

func chainMiddleware(mws []middleware.Middleware, final middleware.Handler) middleware.Handler {
	h := final
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
