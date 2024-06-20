package bo

import (
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
)

type CheckPermissionParams struct {
	JwtClaims *middleware.JwtClaims
	Operation string
}

type CheckTokenParams struct {
	JwtClaims *middleware.JwtClaims
}

type LoginParams struct {
	Username   string
	EnPassword string // 加密后的密码
	Team       uint32 // 对应团队ID
}

type LoginReply struct {
	JwtClaims *middleware.JwtClaims
	User      *model.SysUser
}

type LogoutParams struct {
	JwtClaims *middleware.JwtClaims
}

type RefreshTokenParams struct {
	JwtClaims *middleware.JwtClaims
	Team      uint32 // 对应团队ID
}

type RefreshTokenReply struct {
	JwtClaims *middleware.JwtClaims
	User      *model.SysUser
}
