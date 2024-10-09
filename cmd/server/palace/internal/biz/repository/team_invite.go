package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model"
)

type TeamInvite interface {
	// InviteUser 邀请用户加入团队
	InviteUser(ctx context.Context, params *bo.InviteUserParams) (*model.SysTeamInvite, error)
	// UpdateInviteStatus 更新邀请状态
	UpdateInviteStatus(ctx context.Context, params *bo.UpdateInviteStatusParams) error
	// UserInviteList 受邀请列表
	UserInviteList(ctx context.Context, params *bo.QueryInviteListParams) ([]*model.SysTeamInvite, error)
	// GetInviteUserByUserIdAndType 获取邀请用户信息
	GetInviteUserByUserIdAndType(ctx context.Context, params *bo.InviteUserParams) (*model.SysTeamInvite, error)
	// GetInviteDetail 获取邀请详情
	GetInviteDetail(ctx context.Context, inviteId uint32) (*model.SysTeamInvite, error)
	// DeleteInvite 删除邀请
	DeleteInvite(ctx context.Context, inviteId uint32) error
}
