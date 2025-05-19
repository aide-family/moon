package build

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do/system"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/aide-family/moon/pkg/util/validate"
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
	role = &system.Role{
		CreatorModel: ToCreatorModel(ctx, roleDo),
		Name:         roleDo.GetName(),
		Remark:       roleDo.GetRemark(),
		Status:       roleDo.GetStatus(),
		Users:        ToUsers(ctx, roleDo.GetUsers()),
		Menus:        nil,
	}
	role.WithContext(ctx)
	return role
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
	role = &system.TeamRole{
		TeamModel: ToTeamModel(ctx, roleDo),
		Name:      roleDo.GetName(),
		Remark:    roleDo.GetRemark(),
		Status:    roleDo.GetStatus(),
		Members:   ToTeamMembers(ctx, roleDo.GetMembers()),
		Menus:     nil,
	}
	role.WithContext(ctx)
	return role
}

func ToTeamRoles(ctx context.Context, roles []do.TeamRole) []*system.TeamRole {
	return slices.MapFilter(roles, func(role do.TeamRole) (*system.TeamRole, bool) {
		if validate.IsNil(role) {
			return nil, false
		}
		return ToTeamRole(ctx, role), true
	})
}
