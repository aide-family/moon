package middler

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwtv4 "github.com/golang-jwt/jwt/v4"
	"github.com/redis/go-redis/v9"
	"prometheus-manager/api/perrors"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/util/hash"
)

// AuthClaims jwt claims
type AuthClaims struct {
	ID   uint32 `json:"id"`
	Role string `json:"role"`
	*jwtv4.RegisteredClaims
}

var (
	secret = []byte("secret")
)

var (
	ErrTokenInvalid = perrors.ErrorUnauthorized("请先登录")
	ErrLogout       = perrors.ErrorUnauthorized("令牌已失效, 请重新登录")
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
func SetSecret(s string) {
	secret = []byte(s)
}

// Expire 把token过期掉
func Expire(ctx context.Context, rdsClient *redis.Client, authClaims *AuthClaims) error {
	timeUnix := authClaims.ExpiresAt.Time.Unix()
	if timeUnix <= time.Now().Unix() {
		return nil
	}
	diffTimeUnix := timeUnix - time.Now().Unix()
	// 如果小于1m, 则设置1m
	if diffTimeUnix < 60 {
		diffTimeUnix = 60
	}

	key := consts.UserLogoutKey.Key(authClaims.MD5()).String()
	return rdsClient.Set(ctx, key, authClaims.String(), time.Duration(diffTimeUnix)*time.Second).Err()
}

// IsLogout 判断token是否被logout
func IsLogout(ctx context.Context, rdsClient *redis.Client, authClaims *AuthClaims) error {
	key := consts.UserLogoutKey.Key(authClaims.MD5()).String()
	if rdsClient.Exists(ctx, key).Val() == 1 {
		return ErrLogout
	}
	return nil
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
func IssueToken(id uint32, role string) (string, error) {
	return IssueTokenWithDuration(id, role, time.Hour*24)
}

// IssueTokenWithDuration issue token with duration
func IssueTokenWithDuration(id uint32, role string, duration time.Duration) (string, error) {
	claims := &AuthClaims{
		ID:   id,
		Role: role,
		RegisteredClaims: &jwtv4.RegisteredClaims{
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(duration)),
		},
	}
	token := jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, claims)
	return token.SignedString(secret)
}

// JwtServer jwt server
func JwtServer() middleware.Middleware {
	return jwt.Server(
		func(token *jwtv4.Token) (interface{}, error) {
			return secret, nil
		},
		jwt.WithSigningMethod(jwtv4.SigningMethodHS256),
		jwt.WithClaims(func() jwtv4.Claims {
			return &AuthClaims{}
		}),
	)
}

// MustLogin 必须登录
func MustLogin(cache ...*redis.Client) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (reply interface{}, err error) {
			// 1. 解析jwt
			authClaims, ok := GetAuthClaims(ctx)
			if !ok {
				return nil, ErrTokenInvalid
			}
			// 判断token是否被人为下线
			if len(cache) > 0 && cache[0] != nil {
				client := cache[0]
				if err = IsLogout(ctx, client, authClaims); err != nil {
					return nil, err
				}
			}
			return handler(ctx, req)
		}
	}
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
