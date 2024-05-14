package bo

import (
	"github.com/aide-cloud/moon/cmd/moon/internal/biz/do/model"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
)

type CheckPermissionParams struct {
	JwtClaims *middleware.JwtClaims
}

type CheckTokenParams struct {
	JwtClaims *middleware.JwtClaims
}

type LoginParams struct {
	Username   string
	EnPassword string // 加密后的密码
	Team       uint32 // 对应团队ID
	TeamRole   uint32 // 对应团队角色ID
}

type LoginReply struct {
	JwtClaims *middleware.JwtClaims
	User      *model.SysUser
}
