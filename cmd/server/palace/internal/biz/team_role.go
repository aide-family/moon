package biz

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewTeamRoleBiz 创建团队角色业务
func NewTeamRoleBiz(teamRoleRepo repository.TeamRole) *TeamRoleBiz {
	return &TeamRoleBiz{
		teamRoleRepo: teamRoleRepo,
	}
}

// TeamRoleBiz 团队角色业务
type TeamRoleBiz struct {
	teamRoleRepo repository.TeamRole
}

// CreateTeamRole 创建团队角色
func (b *TeamRoleBiz) CreateTeamRole(ctx context.Context, teamRole *bo.CreateTeamRoleParams) (*bizmodel.SysTeamRole, error) {
	detail, err := b.teamRoleRepo.CreateTeamRole(ctx, teamRole)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return detail, nil
}

// UpdateTeamRole 更新团队角色
func (b *TeamRoleBiz) UpdateTeamRole(ctx context.Context, teamRole *bo.UpdateTeamRoleParams) error {
	if err := b.teamRoleRepo.UpdateTeamRole(ctx, teamRole); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// DeleteTeamRole 删除团队角色
func (b *TeamRoleBiz) DeleteTeamRole(ctx context.Context, id uint32) error {
	if err := b.teamRoleRepo.DeleteTeamRole(ctx, id); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}

// GetTeamRole 获取团队角色
func (b *TeamRoleBiz) GetTeamRole(ctx context.Context, id uint32) (*bizmodel.SysTeamRole, error) {
	role, err := b.teamRoleRepo.GetTeamRole(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nTeamRoleNotFoundErr(ctx)
		}
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return role, nil
}

// ListTeamRole 获取团队角色列表
func (b *TeamRoleBiz) ListTeamRole(ctx context.Context, params *bo.ListTeamRoleParams) ([]*bizmodel.SysTeamRole, error) {
	list, err := b.teamRoleRepo.ListTeamRole(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return list, nil
}

// UpdateTeamRoleStatus 更新团队角色状态
func (b *TeamRoleBiz) UpdateTeamRoleStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	if err := b.teamRoleRepo.UpdateTeamRoleStatus(ctx, status, ids...); !types.IsNil(err) {
		return merr.ErrorI18nSystemErr(ctx).WithCause(err)
	}
	return nil
}
