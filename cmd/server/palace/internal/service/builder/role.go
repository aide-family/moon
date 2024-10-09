package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	teamapi "github.com/aide-family/moon/api/admin/team"
	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ IRoleModuleBuilder = (*roleModuleBuilder)(nil)

type (
	roleModuleBuilder struct {
		ctx context.Context
	}

	IRoleModuleBuilder interface {
		WithCreateRoleRequest(*teamapi.CreateRoleRequest) ICreateRoleRequestBuilder
		WithUpdateRoleRequest(*teamapi.UpdateRoleRequest) IUpdateRoleRequestBuilder
		WithListRoleRequest(*teamapi.ListRoleRequest) IListRoleRequestBuilder
		DoRoleBuilder() IDoRoleBuilder
	}

	ICreateRoleRequestBuilder interface {
		ToBo() *bo.CreateTeamRoleParams
	}

	createRoleRequestBuilder struct {
		ctx context.Context
		*teamapi.CreateRoleRequest
	}

	IUpdateRoleRequestBuilder interface {
		ToBo() *bo.UpdateTeamRoleParams
	}

	updateRoleRequestBuilder struct {
		ctx context.Context
		*teamapi.UpdateRoleRequest
	}

	IListRoleRequestBuilder interface {
		ToBo() *bo.ListTeamRoleParams
	}

	listRoleRequestBuilder struct {
		ctx context.Context
		*teamapi.ListRoleRequest
	}

	IDoRoleBuilder interface {
		ToAPI(*bizmodel.SysTeamRole, ...map[uint32]*adminapi.UserItem) *adminapi.TeamRole
		ToAPIs([]*bizmodel.SysTeamRole) []*adminapi.TeamRole
		ToSelect(*bizmodel.SysTeamRole) *adminapi.SelectItem
		ToSelects([]*bizmodel.SysTeamRole) []*adminapi.SelectItem
	}

	doRoleBuilder struct {
		ctx context.Context
	}
)

func (d *doRoleBuilder) ToAPI(role *bizmodel.SysTeamRole, userMaps ...map[uint32]*adminapi.UserItem) *adminapi.TeamRole {
	if types.IsNil(d) || types.IsNil(role) {
		return nil
	}

	userMap := getUsers(d.ctx, userMaps, role.CreatorID)
	return &adminapi.TeamRole{
		Id:        role.ID,
		Name:      role.Name,
		Remark:    role.Remark,
		CreatedAt: role.CreatedAt.String(),
		UpdatedAt: role.UpdatedAt.String(),
		Status:    api.Status(role.Status),
		Resources: nil, // TODO 关联资源
		Creator:   userMap[role.CreatorID],
	}
}

func (d *doRoleBuilder) ToAPIs(roles []*bizmodel.SysTeamRole) []*adminapi.TeamRole {
	if types.IsNil(d) || types.IsNil(roles) {
		return nil
	}

	ids := types.SliceTo(roles, func(role *bizmodel.SysTeamRole) uint32 {
		return role.CreatorID
	})
	userMap := getUsers(d.ctx, nil, ids...)
	return types.SliceTo(roles, func(role *bizmodel.SysTeamRole) *adminapi.TeamRole {
		return d.ToAPI(role, userMap)
	})
}

func (d *doRoleBuilder) ToSelect(role *bizmodel.SysTeamRole) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(role) {
		return nil
	}

	return &adminapi.SelectItem{
		Value:    role.ID,
		Label:    role.Name,
		Children: nil,
		Disabled: role.DeletedAt > 0 || !role.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: role.Remark,
		},
	}
}

func (d *doRoleBuilder) ToSelects(roles []*bizmodel.SysTeamRole) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(roles) {
		return nil
	}

	return types.SliceTo(roles, func(role *bizmodel.SysTeamRole) *adminapi.SelectItem {
		return d.ToSelect(role)
	})
}

func (l *listRoleRequestBuilder) ToBo() *bo.ListTeamRoleParams {
	if types.IsNil(l) || types.IsNil(l.ListRoleRequest) {
		return nil
	}
	claims, ok := middleware.ParseJwtClaims(l.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(l.ctx))
	}
	return &bo.ListTeamRoleParams{
		TeamID:  claims.GetTeam(),
		Keyword: l.GetKeyword(),
	}
}

func (u *updateRoleRequestBuilder) ToBo() *bo.UpdateTeamRoleParams {
	if types.IsNil(u) || types.IsNil(u.UpdateRoleRequest) {
		return nil
	}

	data := u.GetData()
	return &bo.UpdateTeamRoleParams{
		ID:          u.GetId(),
		Name:        data.GetName(),
		Remark:      data.GetRemark(),
		Permissions: data.GetPermissions(),
	}
}

func (c *createRoleRequestBuilder) ToBo() *bo.CreateTeamRoleParams {
	if types.IsNil(c) || types.IsNil(c.CreateRoleRequest) {
		return nil
	}

	claims, ok := middleware.ParseJwtClaims(c.ctx)
	if !ok {
		panic(merr.ErrorI18nUnauthorized(c.ctx))
	}

	return &bo.CreateTeamRoleParams{
		TeamID:      claims.GetTeam(),
		Name:        c.GetName(),
		Remark:      c.GetRemark(),
		Status:      vobj.StatusEnable,
		Permissions: c.GetPermissions(),
	}
}

func (r *roleModuleBuilder) WithCreateRoleRequest(request *teamapi.CreateRoleRequest) ICreateRoleRequestBuilder {
	return &createRoleRequestBuilder{ctx: r.ctx, CreateRoleRequest: request}
}

func (r *roleModuleBuilder) WithUpdateRoleRequest(request *teamapi.UpdateRoleRequest) IUpdateRoleRequestBuilder {
	return &updateRoleRequestBuilder{ctx: r.ctx, UpdateRoleRequest: request}
}

func (r *roleModuleBuilder) WithListRoleRequest(request *teamapi.ListRoleRequest) IListRoleRequestBuilder {
	return &listRoleRequestBuilder{ctx: r.ctx, ListRoleRequest: request}
}

func (r *roleModuleBuilder) DoRoleBuilder() IDoRoleBuilder {
	return &doRoleBuilder{ctx: r.ctx}
}
