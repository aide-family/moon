// Package middleware is a middleware package for kratos.
package middleware

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/transport"
	jwtv5 "github.com/golang-jwt/jwt/v5"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"github.com/aide-family/moon/cmd/palace/internal/helper/permission"
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/cnst"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/aide-family/moon/pkg/util/validate"
)

// JwtBaseInfo jwt base info
type JwtBaseInfo struct {
	UserID   uint32      `json:"user_id"`
	Username string      `json:"username"`
	Nickname string      `json:"nickname"`
	Avatar   string      `json:"avatar"`
	Gender   vobj.Gender `json:"gender"`
}

// JwtClaims jwt claims
type JwtClaims struct {
	signKey string
	JwtBaseInfo
	jwtv5.RegisteredClaims
}

// ParseJwtClaims parse jwt claims
func ParseJwtClaims(ctx context.Context) (*JwtClaims, bool) {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return nil, false
	}
	jwtClaims, ok := claims.(*JwtClaims)
	if !ok {
		return nil, false
	}
	return jwtClaims, true
}

// ParseJwtClaimsFromToken parse jwt claims from token
func ParseJwtClaimsFromToken(token, signKey string) (*JwtClaims, error) {
	claims, err := jwtv5.Parse(token, func(token *jwtv5.Token) (interface{}, error) {
		return []byte(signKey), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, merr.ErrorInvalidToken("token is invalid")
	}

	claimsBs, err := json.Marshal(claims.Claims)
	if err != nil {
		return nil, err
	}
	var jwtClaims JwtClaims
	if err = json.Unmarshal(claimsBs, &jwtClaims); err != nil {
		return nil, err
	}
	return &jwtClaims, nil
}

// JwtServer jwt server
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

type TokenValidateFunc func(ctx context.Context, token string) (userDo do.User, err error)

// MustLogin must login
func MustLogin(validateFunc TokenValidateFunc) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req any) (any, error) {
			menuDo, ok := do.GetMenuDoContext(ctx)
			if !ok {
				return nil, merr.ErrorBadRequest("not allow request")
			}
			if !menuDo.GetProcessType().IsContainsLogin() {
				return handler(ctx, req)
			}
			claims, ok := ParseJwtClaims(ctx)
			if !ok {
				return nil, merr.ErrorUnauthorized("token error")
			}
			ctx = permission.WithUserIDContext(ctx, claims.UserID)
			tr, ok := transport.FromServerContext(ctx)
			if !ok {
				return nil, merr.ErrorBadRequest("not allow request")
			}
			authorization := tr.RequestHeader().Get(cnst.XHeaderToken)
			if validate.TextIsNull(authorization) {
				return nil, jwt.ErrMissingJwtToken
			}
			auths := strings.SplitN(authorization, " ", 2)
			if len(auths) != 2 || !strings.EqualFold(auths[0], cnst.BearerWord) {
				return nil, jwt.ErrTokenInvalid
			}
			jwtToken := auths[1]
			ctx = permission.WithTokenContext(ctx, jwtToken)
			userDo, err := validateFunc(ctx, jwtToken)
			if err != nil {
				return nil, err
			}
			ctx = do.WithUserDoContext(ctx, userDo)
			return handler(ctx, req)
		}
	}
}

// NewJwtClaims new jwt claims
func NewJwtClaims(c *config.JWT, base JwtBaseInfo) *JwtClaims {
	return &JwtClaims{
		signKey:     c.GetSignKey(),
		JwtBaseInfo: base,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(timex.Now().Add(c.GetExpire().AsDuration())),
			Issuer:    c.GetIssuer(),
		},
	}
}

// GetToken get token
func (l *JwtClaims) GetToken() (string, error) {
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, l).SignedString([]byte(l.signKey))
}
