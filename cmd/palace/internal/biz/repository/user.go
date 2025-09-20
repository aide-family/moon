package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

type User interface {
	FindByID(ctx context.Context, userID uint32) (do.User, error)
	FindByEmail(ctx context.Context, email string) (do.User, error)
	CreateUserWithOAuthUser(ctx context.Context, user bo.IOAuthUser, sendEmailFunc bo.SendEmailFun) (do.User, error)
	SetEmail(ctx context.Context, user do.User, sendEmailFunc bo.SendEmailFun) (do.User, error)
	Create(ctx context.Context, user do.User, sendEmailFunc bo.SendEmailFun) (do.User, error)
	UpdateUserInfo(ctx context.Context, user do.User) error
	UpdateUserAvatar(ctx context.Context, userID uint32, avatar string) error
	UpdateUserPhone(ctx context.Context, userID uint32, phone string) error
	UpdateUserEmail(ctx context.Context, userID uint32, email string) error
	UpdatePassword(ctx context.Context, updateUserPasswordInfo *bo.UpdateUserPasswordInfo) error
	GetTeamsByUserID(ctx context.Context, userID uint32) ([]do.Team, error)
	AppendTeam(ctx context.Context, team do.Team) error
	UpdateUserStatus(ctx context.Context, req *bo.UpdateUserStatusRequest) error
	UpdateUserPosition(ctx context.Context, req bo.UpdateUserPosition) error
	List(ctx context.Context, req *bo.UserListRequest) (*bo.UserListReply, error)
	Find(ctx context.Context, ids []uint32) ([]do.User, error)
	UpdateUserRoles(ctx context.Context, req bo.UpdateUserRoles) error
	Get(ctx context.Context, id uint32) (do.User, error)
}
