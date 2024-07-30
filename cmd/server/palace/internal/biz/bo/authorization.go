package bo

import (
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
)

// CheckPermissionParams 鉴权请求参数
type CheckPermissionParams struct {
	JwtClaims *middleware.JwtClaims
	Operation string
}

// CheckTokenParams token鉴权请求参数
type CheckTokenParams struct {
	JwtClaims *middleware.JwtClaims
}

// LoginParams 登录请求参数
type LoginParams struct {
	Username   string
	EnPassword string // 加密后的密码
	Team       uint32 // 对应团队ID
}

// LoginReply 登录响应
type LoginReply struct {
	JwtClaims *middleware.JwtClaims
	User      *model.SysUser
}

// LogoutParams 登出请求参数
type LogoutParams struct {
	JwtClaims *middleware.JwtClaims
}

// RefreshTokenParams 刷新token请求参数
type RefreshTokenParams struct {
	JwtClaims *middleware.JwtClaims
	Team      uint32 // 对应团队ID
}

// RefreshTokenReply 刷新token响应
type RefreshTokenReply struct {
	JwtClaims *middleware.JwtClaims
	User      *model.SysUser
}
