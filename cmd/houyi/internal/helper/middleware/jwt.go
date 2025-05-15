package middleware

import (
	"github.com/aide-family/moon/pkg/config"
	"github.com/aide-family/moon/pkg/util/timex"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

func NewJwtClaims(c *config.JWT, token string) *JwtClaims {
	now := timex.Now()
	return &JwtClaims{
		signKey: c.SignKey,
		Token:   token,
		RegisteredClaims: &jwtv5.RegisteredClaims{
			Issuer:    c.Issuer,
			Subject:   "moon.houyi",
			ExpiresAt: jwtv5.NewNumericDate(now.Add(c.GetExpire().AsDuration())),
			IssuedAt:  jwtv5.NewNumericDate(now),
			NotBefore: jwtv5.NewNumericDate(now),
		},
	}
}

// JwtClaims jwt claims
type JwtClaims struct {
	signKey string
	Token   string `json:"token"`
	Name    string `json:"name"`
	*jwtv5.RegisteredClaims
}

// GetToken get token
func (l *JwtClaims) GetToken() (string, error) {
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, l).SignedString([]byte(l.signKey))
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
