// Package jwt provides a JWT token generator and parser.
package jwt

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/strutil"
	"github.com/bwmarrin/snowflake"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

type (
	// BaseInfo is the base information of the JWT token.
	BaseInfo struct {
		UID      snowflake.ID `json:"uid"`
		Username string       `json:"username"`
	}

	// JwtClaims is the claims of the JWT token.
	JwtClaims struct {
		signKey string
		BaseInfo
		jwtv5.RegisteredClaims
	}
)

// NewJwtClaims creates a new JWT claims.
// If the JWT config is not valid, it will return an error.
func NewJwtClaims(c *config.JWT, base BaseInfo) *JwtClaims {
	expire, issuer := c.GetExpire().AsDuration(), c.GetIssuer()
	if expire <= 0 {
		expire = 10 * time.Minute
	}
	if strutil.IsEmpty(issuer) {
		issuer = "moon"
	}
	return &JwtClaims{
		signKey:  c.GetSecret(),
		BaseInfo: base,
		RegisteredClaims: jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(expire)),
			Issuer:    issuer,
		},
	}
}

// GenerateToken generates a new JWT token.
// If the JWT token is not valid, it will return an error.
func (l *JwtClaims) GenerateToken() (string, error) {
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, l).SignedString([]byte(l.signKey))
}

// GetClaimsFromContext gets the JWT claims from the context.
// If the JWT claims are not valid, it will return an error.
func GetClaimsFromContext(ctx context.Context) (*JwtClaims, error) {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return nil, merr.ErrorUnauthorized("token is required")
	}
	jwtClaims, ok := claims.(*JwtClaims)
	if !ok {
		return nil, merr.ErrorUnauthorized("token is invalid")
	}
	return jwtClaims, nil
}

// ParseClaimsFromToken parses the JWT claims from the token string.
// If the JWT claims are not valid, it will return an error.
func ParseClaimsFromToken(secret string, token string) (*JwtClaims, error) {
	claims, err := jwtv5.Parse(token, func(token *jwtv5.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if !claims.Valid {
		return nil, merr.ErrorUnauthorized("token is invalid")
	}

	claimsBs, err := json.Marshal(claims.Claims)
	if err != nil {
		return nil, err
	}
	var jwtClaims JwtClaims
	if err := json.Unmarshal(claimsBs, &jwtClaims); err != nil {
		return nil, err
	}
	return &jwtClaims, nil
}
