package biz

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repo"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewTeamRoleBiz(teamRoleRepo repo.TeamRoleRepo) *TeamRoleBiz {
	return &TeamRoleBiz{
		teamRoleRepo: teamRoleRepo,
	}
}

type TeamRoleBiz struct {
	teamRoleRepo repo.TeamRoleRepo
}

// CreateTeamRole 创建团队角色
func (b *TeamRoleBiz) CreateTeamRole(ctx context.Context, teamRole *bo.CreateTeamRoleParams) (*model.SysTeamRole, error) {
	return b.teamRoleRepo.CreateTeamRole(ctx, teamRole)
}

// UpdateTeamRole 更新团队角色
func (b *TeamRoleBiz) UpdateTeamRole(ctx context.Context, teamRole *bo.UpdateTeamRoleParams) error {
	return b.teamRoleRepo.UpdateTeamRole(ctx, teamRole)
}

// DeleteTeamRole 删除团队角色
func (b *TeamRoleBiz) DeleteTeamRole(ctx context.Context, id uint32) error {
	return b.teamRoleRepo.DeleteTeamRole(ctx, id)
}

// GetTeamRole 获取团队角色
func (b *TeamRoleBiz) GetTeamRole(ctx context.Context, id uint32) (*model.SysTeamRole, error) {
	return b.teamRoleRepo.GetTeamRole(ctx, id)
}

// ListTeamRole 获取团队角色列表
func (b *TeamRoleBiz) ListTeamRole(ctx context.Context, params *bo.ListTeamRoleParams) ([]*model.SysTeamRole, error) {
	return b.teamRoleRepo.ListTeamRole(ctx, params)
}

// UpdateTeamRoleStatus 更新团队角色状态
func (b *TeamRoleBiz) UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	return b.teamRoleRepo.UpdateTeamRoleStatus(ctx, status, ids...)
}
