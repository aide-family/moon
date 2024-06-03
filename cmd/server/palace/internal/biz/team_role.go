package biz

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

func NewTeamRoleBiz(teamRoleRepo repository.TeamRole) *TeamRoleBiz {
	return &TeamRoleBiz{
		teamRoleRepo: teamRoleRepo,
	}
}

type TeamRoleBiz struct {
	teamRoleRepo repository.TeamRole
}

// CreateTeamRole 创建团队角色
func (b *TeamRoleBiz) CreateTeamRole(ctx context.Context, teamRole *bo.CreateTeamRoleParams) (*bizmodel.SysTeamRole, error) {
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
func (b *TeamRoleBiz) GetTeamRole(ctx context.Context, id uint32) (*bizmodel.SysTeamRole, error) {
	role, err := b.teamRoleRepo.GetTeamRole(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, bo.TeamRoleNotFoundErr(ctx)
		}
		return nil, err
	}
	return role, nil
}

// ListTeamRole 获取团队角色列表
func (b *TeamRoleBiz) ListTeamRole(ctx context.Context, params *bo.ListTeamRoleParams) ([]*bizmodel.SysTeamRole, error) {
	return b.teamRoleRepo.ListTeamRole(ctx, params)
}

// UpdateTeamRoleStatus 更新团队角色状态
func (b *TeamRoleBiz) UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	return b.teamRoleRepo.UpdateTeamRoleStatus(ctx, status, ids...)
}
