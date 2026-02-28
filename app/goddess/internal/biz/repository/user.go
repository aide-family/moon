package repository

import (
	"context"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/goddess/internal/biz/bo"
)

type User interface {
	GetUser(ctx context.Context, uid snowflake.ID) (*bo.UserItemBo, error)
	ListUser(ctx context.Context, req *bo.ListUserBo) (*bo.PageResponseBo[*bo.UserItemBo], error)
	SelectUser(ctx context.Context, req *bo.SelectUserBo) (*bo.SelectUserBoResult, error)
	GetUserByEmail(ctx context.Context, email string) (*bo.UserItemBo, error)
	UpdateUserStatus(ctx context.Context, uid snowflake.ID, status enum.UserStatus) error
	UpdateUserEmail(ctx context.Context, uid snowflake.ID, email string) error
	UpdateUserAvatar(ctx context.Context, uid snowflake.ID, avatar string) error
}
