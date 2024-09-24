package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo/auth"
	"github.com/aide-family/moon/pkg/palace/model"
)

// GithubUser .
type GithubUser interface {
	// FirstOrCreate 获取用户信息， 如果没有则创建
	FirstOrCreate(context.Context, *auth.GithubLoginResponse) (*model.SysUser, error)
}
