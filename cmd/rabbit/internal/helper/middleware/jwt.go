// Package middleware is a middleware package for kratos.
package middleware

import (
	"context"
	"strconv"

	"github.com/aide-family/moon/cmd/rabbit/internal/helper/permission"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

// NewJwtClaims creates a new JwtClaims.
func NewJwtClaims(c *config.JWT, token string) *JwtClaims {
	now := timex.Now()
	return &JwtClaims{
		signKey: c.SignKey,
		Token:   token,
		RegisteredClaims: jwtv5.RegisteredClaims{
			Issuer:    c.Issuer,
			Subject:   "moon.rabbit",
			ExpiresAt: jwtv5.NewNumericDate(now.Add(c.GetExpire().AsDuration())),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
		},
	}
}

// JwtClaims is a jwt claims.
type JwtClaims struct {
	signKey string
	Token   string `json:"token"`
	Name    string `json:"name"`
	jwtv5.RegisteredClaims
}

// GetToken gets the token.
func (l *JwtClaims) GetToken() (string, error) {
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, l).SignedString([]byte(l.signKey))
}

// JwtServer is a jwt server.
func JwtServer(signKey string) middleware.Middleware {
	return jwt.Server(
		func(token *jwtv5.Token) (interface{}, error) {
			return []byte(signKey), nil
		},
		jwt.WithSigningMethod(jwtv5.SigningMethodHS256),
		jwt.WithClaims(func() jwtv5.Claims {
			return &JwtClaims{}
		}),
	)
}

// ExtractMetadata extracts the teamId and token from the metadata or HTTP header and sets them to the context.
func ExtractMetadata() middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			// First try to get the teamId from the metadata.
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

			// If the metadata does not have the teamId, try to get it from the HTTP header.
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
