package build

import (
	"context"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do/system"
	"github.com/moon-monitor/moon/pkg/util/slices"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

func ToRole(ctx context.Context, roleDo do.Role) *system.Role {
	if validate.IsNil(roleDo) {
		return nil
	}
	role, ok := roleDo.(*system.Role)
	if ok {
		role.WithContext(ctx)
		return role
	}
	return &system.Role{
		CreatorModel: ToCreatorModel(ctx, roleDo),
		Name:         role.GetName(),
		Remark:       role.GetRemark(),
		Status:       role.GetStatus(),
		Users:        ToUsers(ctx, role.GetUsers()),
		Menus:        nil,
	}
}

func ToRoles(ctx context.Context, roles []do.Role) []*system.Role {
	return slices.MapFilter(roles, func(role do.Role) (*system.Role, bool) {
		if validate.IsNil(role) {
			return nil, false
		}
		return ToRole(ctx, role), true
	})
}

func ToTeamRole(ctx context.Context, roleDo do.TeamRole) *system.TeamRole {
	if validate.IsNil(roleDo) {
		return nil
	}
	role, ok := roleDo.(*system.TeamRole)
	if ok {
		role.WithContext(ctx)
		return role
	}
	return &system.TeamRole{
		TeamModel: ToTeamModel(ctx, roleDo),
		Name:      role.GetName(),
		Remark:    role.GetRemark(),
		Status:    role.GetStatus(),
		Members:   ToTeamMembers(ctx, role.GetMembers()),
		Menus:     nil,
	}
}

func ToTeamRoles(ctx context.Context, roles []do.TeamRole) []*system.TeamRole {
	return slices.MapFilter(roles, func(role do.TeamRole) (*system.TeamRole, bool) {
		if validate.IsNil(role) {
			return nil, false
		}
		return ToTeamRole(ctx, role), true
	})
}
