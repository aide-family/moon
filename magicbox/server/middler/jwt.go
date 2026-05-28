// Package middler is a package for middleware.
package middler

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/pointer"
	"github.com/aide-family/magicbox/strutil"
	"github.com/aide-family/magicbox/strutil/cnst"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/middleware"
	kjwt "github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/aide-family/magicbox/contextx"
	"github.com/aide-family/magicbox/jwt"
)

// JwtClient propagates JWT auth for downstream calls. Prefer AuthClient for exclusive JWT/service-key routing.
func JwtClient(headers ...string) middleware.Middleware {
	return AuthClient(headers...)
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
			fullAuth, _, ok := authorizationFromContext(ctx)
			if !ok {
				return nil, merr.ErrorUnauthorized("token is invalid")
			}
			tr.RequestHeader().Set(cnst.HTTPHeaderAuthorization, fullAuth)
			tr.RequestHeader().Set(cnst.MetadataGlobalKeyAuthorization, fullAuth)
			return handler(ctx, req)
		}
	}
}

func ValidateUser(validator func(ctx context.Context, userUID snowflake.ID) error) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			userUID := contextx.GetUserUID(ctx)
			if userUID == 0 {
				return nil, merr.ErrorUnauthorized("user is not authenticated")
			}
			if err := validator(ctx, userUID); err != nil {
				return nil, err
			}
			return handler(ctx, req)
		}
	}
}
