// Package middler is a package for middleware.
package middler

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/jwt"
)

func JwtClient(headers ...string) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			clientContext, ok := transport.FromClientContext(ctx)
			if !ok {
				return handler(ctx, req)
			}
			if tr, ok := transport.FromServerContext(ctx); ok {
				if auth := tr.RequestHeader().Get(cnst.HTTPHeaderAuthorization); strutil.IsNotEmpty(auth) {
					clientContext.RequestHeader().Set(cnst.HTTPHeaderAuthorization, auth)
				}
				if namespace := tr.RequestHeader().Get(cnst.HTTPHeaderXNamespace); strutil.IsNotEmpty(namespace) {
					clientContext.RequestHeader().Set(cnst.HTTPHeaderXNamespace, namespace)
				}
				for _, header := range headers {
					if value := tr.RequestHeader().Get(header); strutil.IsNotEmpty(value) {
						clientContext.RequestHeader().Set(header, value)
					}
				}
			}
			if md, ok := metadata.FromClientContext(ctx); ok {
				if auth := md.Get(cnst.MetadataGlobalKeyAuthorization); strutil.IsNotEmpty(auth) {
					clientContext.RequestHeader().Set(cnst.HTTPHeaderAuthorization, auth)
				}
				if namespace := md.Get(cnst.MetadataGlobalKeyNamespace); strutil.IsNotEmpty(namespace) {
					clientContext.RequestHeader().Set(cnst.HTTPHeaderXNamespace, namespace)
				}
				for _, header := range headers {
					if value := md.Get(header); strutil.IsNotEmpty(value) {
						clientContext.RequestHeader().Set(header, value)
					}
				}
			}

			return handler(ctx, req)
		}
	}
}

func JwtServe(signKey string, claims jwtv5.Claims) middleware.Middleware {
	return kjwt.Server(
		func(token *jwtv5.Token) (interface{}, error) {
			return []byte(signKey), nil
		},
		kjwt.WithSigningMethod(jwtv5.SigningMethodHS256),
		kjwt.WithClaims(func() jwtv5.Claims {
			return claims
		}),
	)
}

func MustLogin() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			claims, err := jwt.GetClaimsFromContext(ctx)
			if err != nil {
				return nil, err
			}
			if pointer.IsNil(claims) || claims.UID == 0 || strutil.IsEmpty(claims.Username) {
				return nil, merr.ErrorUnauthorized("token is invalid")
			}
			ctx = contextx.WithUserUID(ctx, claims.UID)
			ctx = contextx.WithUsername(ctx, claims.Username)
			return handler(ctx, req)
		}
	}
}

func BindJwtToken() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, merr.ErrorUnauthorized("wrong context for middleware")
			}
			authToken := tr.RequestHeader().Get(cnst.HTTPHeaderAuthorization)
			auths := strings.SplitN(tr.RequestHeader().Get(cnst.HTTPHeaderAuthorization), " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], cnst.HTTPHeaderBearerPrefix) {
				return nil, merr.ErrorUnauthorized("token is invalid")
			}

			tr.RequestHeader().Set(cnst.MetadataGlobalKeyAuthorization, authToken)
			return handler(ctx, req)
		}
	}
}
