package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo/auth"
	"github.com/aide-family/moon/pkg/palace/model"
)

// OAuth .
type OAuth interface {
	// OAuthUserFirstOrCreate 获取用户信息， 如果没有则创建
	OAuthUserFirstOrCreate(context.Context, auth.IOAuthUser) (*model.SysUser, error)
}
