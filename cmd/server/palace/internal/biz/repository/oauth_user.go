package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo/auth"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

// OAuth .
type OAuth interface {
	// OAuthUserFirstOrCreate 获取用户信息， 如果没有则创建
	OAuthUserFirstOrCreate(context.Context, auth.IOAuthUser) (*model.SysUser, error)

	// SetEmail 设置电子邮箱
	SetEmail(context.Context, string, string) (*model.SysUser, error)

	GetSysUserByOAuthID(context.Context, string, vobj.OAuthAPP) (*model.SysOAuthUser, error)

	SendVerifyEmail(ctx context.Context, email string) error

	CheckVerifyEmailCode(ctx context.Context, email, code string) error
}
