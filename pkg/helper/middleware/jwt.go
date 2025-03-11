package middleware

import (
	"context"
	"strconv"
	"sync"
	"time"

	"github.com/aide-family/moon/api/admin/authorization"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
	"google.golang.org/protobuf/types/known/durationpb"

	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/auth/jwt"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	jwtv5 "github.com/golang-jwt/jwt/v5"
)

var (
	signKey = "moon-monitor"
	issuer  = "moon-monitor"
	expire  = time.Hour * 24 // 默认24小时

	signKeyOnce sync.Once
	expireOnce  sync.Once
	issuerOnce  sync.Once
)

// JWTConfig jwt config
type JWTConfig interface {
	// GetSignKey get sign key
	GetSignKey() string
	// GetExpire get expire
	GetExpire() *durationpb.Duration
	// GetIssuer get issuer
	GetIssuer() string
}

// SetJwtConfig set jwt config
func SetJwtConfig(cfg JWTConfig) {
	SetSignKey(cfg.GetSignKey())
	SetExpire(cfg.GetExpire().AsDuration())
	SetIssuer(cfg.GetIssuer())
}

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
	*jwtv5.RegisteredClaims
}

// JwtBaseInfo jwt base info
type JwtBaseInfo struct {
	UserID uint32 `json:"user"`
}

// GetUser 获取用户id
func (l *JwtBaseInfo) GetUser() uint32 {
	if types.IsNil(l) {
		return 0
	}
	return l.UserID
}

// SetUserInfo 设置用户信息
func (l *JwtBaseInfo) SetUserInfo(userID uint32) *JwtBaseInfo {
	l.UserID = userID
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

// ParseJwtClaimsFromToken parse jwt claims from token
func ParseJwtClaimsFromToken(token string) (*JwtBaseInfo, bool) {
	claims, err := jwtv5.Parse(token, func(token *jwtv5.Token) (interface{}, error) {
		return []byte(signKey), nil
	})
	if err != nil {
		log.Errorw("解析token失败：", err, "token", token)
		return nil, false
	}
	if !claims.Valid {
		log.Errorw("token无效：", "token", token)
		return nil, false
	}

	claimsBs, _ := types.Marshal(claims.Claims)
	log.Debugw("claims", string(claimsBs))
	var jwtClaims JwtClaims
	err = types.Unmarshal(claimsBs, &jwtClaims)
	if err != nil {
		log.Errorw("解析token失败：", err, "token", token)
		return nil, false
	}
	return jwtClaims.JwtBaseInfo, true
}

// NewJwtClaims new jwt claims
func NewJwtClaims(base *JwtBaseInfo) *JwtClaims {
	return &JwtClaims{
		JwtBaseInfo: base,
		RegisteredClaims: &jwtv5.RegisteredClaims{
			ExpiresAt: jwtv5.NewNumericDate(time.Now().Add(expire)),
			Issuer:    issuer,
		},
	}
}

// GetToken get token
func (l *JwtClaims) GetToken() (string, error) {
	return jwtv5.NewWithClaims(jwtv5.SigningMethodHS256, l).SignedString([]byte(signKey))
}

// Logout 缓存token hash
func (l *JwtClaims) Logout(ctx context.Context, cache cache.ICacher) error {
	token, err := l.GetToken()
	if err != nil {
		return err
	}
	bs, _ := types.Marshal(l)
	return cache.Client().Set(ctx, types.MD5(token), string(bs), expire).Err()
}

// IsLogout 是否已经登出
func (l *JwtClaims) IsLogout(ctx context.Context, cache cache.ICacher) bool {
	return isLogout(ctx, cache, l)
}

type (
	userRoleContextKey     struct{}
	userIDContextKey       struct{}
	teamIDContextKey       struct{}
	teamMemberIDContextKey struct{}
)

// WithUserRoleContextKey with user role context key
func WithUserRoleContextKey(ctx context.Context, role vobj.Role) context.Context {
	return context.WithValue(ctx, userRoleContextKey{}, role)
}

// WithUserIDContextKey with user id context key
func WithUserIDContextKey(ctx context.Context, id uint32) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, id)
}

// WithTeamIDContextKey with team id context key
func WithTeamIDContextKey(ctx context.Context, id uint32) context.Context {
	return context.WithValue(ctx, teamIDContextKey{}, id)
}

// WithTeamMemberIDContextKey with team member id context key
func WithTeamMemberIDContextKey(ctx context.Context, id uint32) context.Context {
	return context.WithValue(ctx, teamMemberIDContextKey{}, id)
}

// GetUserID get user id
func GetUserID(ctx context.Context) uint32 {
	id, ok := ctx.Value(userIDContextKey{}).(uint32)
	if !ok {
		return 0
	}
	return id
}

// GetTeamID get team id
func GetTeamID(ctx context.Context) uint32 {
	id, ok := ctx.Value(teamIDContextKey{}).(uint32)
	if !ok {
		return 0
	}
	return id
}

// GetTeamMemberID get team member id
func GetTeamMemberID(ctx context.Context) uint32 {
	id, ok := ctx.Value(teamMemberIDContextKey{}).(uint32)
	if !ok {
		return 0
	}
	return id
}

// GetUserRole get user role
func GetUserRole(ctx context.Context) vobj.Role {
	role, ok := ctx.Value(userRoleContextKey{}).(vobj.Role)
	if !ok {
		return vobj.RoleUser
	}
	return role
}

// CheckTokenFun check token fun
type CheckTokenFun func(ctx context.Context) (*authorization.CheckTokenReply, error)

const (
	XTeamIDHeader       = "X-Team-ID"
	XTeamMemberIDHeader = "X-Team-Member-ID"
)

// JwtLoginMiddleware jwt login middleware
func JwtLoginMiddleware(check CheckTokenFun) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			checked, err := check(ctx)
			if err != nil {
				return nil, err
			}
			if !checked.GetIsLogin() {
				return nil, merr.ErrorI18nUnauthorized(ctx)
			}
			if tr, ok := transport.FromServerContext(ctx); ok {
				if teamIDStr := tr.RequestHeader().Get(XTeamIDHeader); teamIDStr != "" {
					teamID, err := strconv.ParseUint(teamIDStr, 10, 32)
					if err != nil {
						return nil, merr.ErrorI18nUnauthorized(ctx)
					}
					ctx = WithTeamIDContextKey(ctx, uint32(teamID))
				}
				if teamMemberIDStr := tr.RequestHeader().Get(XTeamMemberIDHeader); teamMemberIDStr != "" {
					teamMemberID, err := strconv.ParseUint(teamMemberIDStr, 10, 32)
					if err != nil {
						return nil, merr.ErrorI18nUnauthorized(ctx)
					}
					ctx = WithTeamMemberIDContextKey(ctx, uint32(teamMemberID))
				}
			}
			ctx = WithUserRoleContextKey(ctx, vobj.Role(checked.GetUser().GetRole()))
			claims, ok := ParseJwtClaims(ctx)
			if !ok {
				return nil, merr.ErrorI18nUnauthorized(ctx)
			}
			ctx = WithUserIDContextKey(ctx, claims.GetUser())
			return handler(ctx, req)
		}
	}
}

// isLogout 是否已经登出
func isLogout(ctx context.Context, cache cache.ICacher, jwtClaims *JwtClaims) bool {
	// 判断是否过期
	token, err := jwtClaims.GetToken()
	if err != nil {
		return true
	}
	exist, err := cache.Client().Exists(ctx, types.MD5(token)).Result()
	return exist == 1 && types.IsNil(err)
}

// IsExpire 是否过期
func IsExpire(jwtClaims *JwtClaims) bool {
	expirationTime, err := jwtClaims.GetExpirationTime()
	if err != nil {
		return true
	}
	// 判断是否过期
	return expirationTime.Before(time.Now())
}

// JwtServer jwt server
func JwtServer() middleware.Middleware {
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

// NewWhiteListMatcher new white list matcher
func NewWhiteListMatcher(list []string) selector.MatchFunc {
	whiteList := make(map[string]struct{})
	for _, v := range list {
		whiteList[v] = struct{}{}
	}
	return func(ctx context.Context, operation string) bool {
		_, ok := whiteList[operation]
		return !ok
	}
}
