package helper

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"

	"prometheus-manager/pkg/util/hash"
)

// AuthClaims jwt claims
type AuthClaims struct {
	ID   uint   `json:"id"`
	Role string `json:"role"`
	*jwtv4.RegisteredClaims
}

var (
	secret = []byte("secret")
)

var (
	ErrTokenInvalid = jwt.ErrTokenInvalid
)

func (l *AuthClaims) MD5() string {
	return hash.MD5(l.String())
}

func (l *AuthClaims) String() string {
	if l == nil {
		return "{}"
	}
	jsonByte, _ := json.Marshal(l)
	return string(jsonByte)
}

// SetSecret set secret
func SetSecret(s []byte) {
	secret = s
}

// Expire 把token过期掉
func Expire(ctx context.Context, rdsClient *redis.Client, authClaims *AuthClaims) error {
	timeUnix := authClaims.ExpiresAt.Time.Unix()
	if timeUnix <= time.Now().Unix() {
		return nil
	}

	return rdsClient.Set(ctx, authClaims.MD5(), authClaims.String(), time.Duration(timeUnix-time.Now().Unix())).Err()
}

// GetAuthClaims get auth claims
func GetAuthClaims(ctx context.Context) (*AuthClaims, bool) {
	claims, ok := jwt.FromContext(ctx)
	if !ok {
		return nil, false
	}
	authClaims, ok := claims.(*AuthClaims)
	if !ok {
		return nil, false
	}

	return authClaims, true
}

// IssueToken issue token
func IssueToken(id uint, role string) (string, error) {
	claims := &AuthClaims{
		ID:   id,
		Role: role,
		RegisteredClaims: &jwtv4.RegisteredClaims{
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// JwtServer jwt server
func JwtServer() middleware.Middleware {
	return jwt.Server(func(token *jwtv4.Token) (interface{}, error) {
		return secret, nil
	}, jwt.WithSigningMethod(jwtv4.SigningMethodHS256),
		jwt.WithClaims(func() jwtv4.Claims {
			return &AuthClaims{}
		}),
	)
}
func NewWhiteListMatcher(list []string) selector.MatchFunc {
	whiteList := make(map[string]struct{})
	for _, v := range list {
		whiteList[v] = struct{}{}
	}
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}
