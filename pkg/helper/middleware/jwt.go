package middleware

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/pkg/conn"
	"github.com/aide-cloud/moon/pkg/utils/cipher"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwtv4 "github.com/golang-jwt/jwt/v4"

	"github.com/aide-cloud/moon/pkg/types"
)

var (
	signKey = "moon-monitor"
	issuer  = "moon-monitor"
	expire  = time.Hour * 24 // 默认24小时

	signKeyOnce sync.Once
	expireOnce  sync.Once
	issuerOnce  sync.Once
)

// SetSignKey set sign key
func SetSignKey(key string) {
	signKeyOnce.Do(func() {
		signKey = key
	})
}

// SetIssuer set issuer
func SetIssuer(iss string) {
	issuerOnce.Do(func() {
		issuer = iss
	})
}

// SetExpire set expire
func SetExpire(e time.Duration) {
	expireOnce.Do(func() {
		expire = e
	})
}

// JwtClaims jwt claims
type JwtClaims struct {
	*JwtBaseInfo
	*jwtv4.RegisteredClaims
}

type JwtBaseInfo struct {
	User     uint32 `json:"user"`
	Role     uint32 `json:"role"`
	Team     uint32 `json:"team"`
	TeamRole uint32 `json:"team_role"`
}

func (l *JwtBaseInfo) GetUser() uint32 {
	if types.IsNil(l) {
		return 0
	}
	return l.User
}

func (l *JwtBaseInfo) GetRole() uint32 {
	if types.IsNil(l) {
		return 0
	}
	return l.Role
}

func (l *JwtBaseInfo) GetTeam() uint32 {
	if types.IsNil(l) {
		return 0
	}
	return l.Team
}

func (l *JwtBaseInfo) GetTeamRole() uint32 {
	if types.IsNil(l) {
		return 0
	}
	return l.TeamRole
}

// SetUserInfo 设置用户信息
func (l *JwtBaseInfo) SetUserInfo(f func() (userId, role uint32, err error)) *JwtBaseInfo {
	userId, role, err := f()
	if err == nil {
		l.User = userId
		l.Role = role
	}

	return l
}

// SetTeamInfo 设置团队信息
func (l *JwtBaseInfo) SetTeamInfo(f func() (teamId, teamRole uint32, err error)) *JwtBaseInfo {
	teamId, teamRole, err := f()
	if err == nil {
		l.Team = teamId
		l.TeamRole = teamRole
	}

	return l
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

// NewJwtClaims new jwt claims
func NewJwtClaims(base *JwtBaseInfo) *JwtClaims {
	return &JwtClaims{
		JwtBaseInfo: base,
		RegisteredClaims: &jwtv4.RegisteredClaims{
			ExpiresAt: jwtv4.NewNumericDate(time.Now().Add(expire)),
			Issuer:    issuer,
		},
	}
}

// GetToken get token
func (l *JwtClaims) GetToken() (string, error) {
	return jwtv4.NewWithClaims(jwtv4.SigningMethodHS256, l).SignedString([]byte(signKey))
}

// Cache 缓存token hash
func (l *JwtClaims) Cache(ctx context.Context, cache conn.Cache) error {
	token, err := l.GetToken()
	if err != nil {
		return err
	}
	bs, _ := json.Marshal(l)
	return cache.Set(ctx, cipher.MD5(token), string(bs), expire)
}

type CheckTokenFun func(ctx context.Context) (bool, error)

// JwtLoginMiddleware jwt login middleware
func JwtLoginMiddleware(check CheckTokenFun) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			isLogin, err := check(ctx)
			if err != nil {
				return nil, err
			}
			if !isLogin {
				return nil, merr.ErrorRedirect("请先登录").WithMetadata(map[string]string{
					"redirect": "/login",
				})
			}
			return handler(ctx, req)
		}
	}
}

// isLogout 是否已经登出
func isLogout(ctx context.Context, cache conn.Cache, jwtClaims *JwtClaims) bool {
	// 判断是否过期
	token, err := jwtClaims.GetToken()
	if err != nil {
		return true
	}
	return cache.Exist(ctx, cipher.MD5(token))
}

// 是否过期
func isExpire(jwtClaims *JwtClaims) bool {
	// 判断是否过期
	return jwtClaims.VerifyExpiresAt(time.Now(), true)
}

// JwtServer jwt server
func JwtServer() middleware.Middleware {
	return jwt.Server(
		func(token *jwtv4.Token) (interface{}, error) {
			return signKey, nil
		},
		jwt.WithSigningMethod(jwtv4.SigningMethodHS256),
		jwt.WithClaims(func() jwtv4.Claims {
			return &JwtClaims{}
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
