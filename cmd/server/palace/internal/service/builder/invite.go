package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	inviteapi "github.com/aide-family/moon/api/admin/invite"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ InviteModuleBuilder = (*inviteModuleBuilder)(nil)

type (
	// InviteModuleBuilder 邀请模块构造器
	InviteModuleBuilder interface {
		// WithInviteUserRequest 创建邀请用户请求参数构造器
		WithInviteUserRequest(*inviteapi.InviteUserRequest) ICreateInviteUserRequestBuilder
		// WithUpdateInviteStatusRequest 更新邀请状态请求参数构造器
		WithUpdateInviteStatusRequest(*inviteapi.UpdateInviteStatusRequest) IUpdateInviteStatusRequestBuilder
		// WithListInviteUserRequest 获取邀请用户列表请求参数构造器
		WithListInviteUserRequest(*inviteapi.ListUserInviteRequest) IListInviteUserRequestBuilder
		// DoInviteBuilder 邀请条目构造器
		DoInviteBuilder(*bo.InviteTeamInfoParams) IDoInviteBuilder
	}

	inviteModuleBuilder struct {
		ctx context.Context
	}

	// ICreateInviteUserRequestBuilder 创建邀请用户请求参数构造器
	ICreateInviteUserRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.InviteUserParams
	}

	createInviteUserRequestBuilder struct {
		ctx context.Context
		*inviteapi.InviteUserRequest
	}

	// IUpdateInviteStatusRequestBuilder 更新邀请状态请求参数构造器
	IUpdateInviteStatusRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.UpdateInviteStatusParams
	}
	updateInviteStatusRequestBuilder struct {
		ctx context.Context
		*inviteapi.UpdateInviteStatusRequest
	}

	// IListInviteUserRequestBuilder 获取邀请用户列表请求参数构造器
	IListInviteUserRequestBuilder interface {
		// ToBo 转换为业务对象
		ToBo() *bo.QueryInviteListParams
	}

	listInviteUserRequestBuilder struct {
		ctx context.Context
		*inviteapi.ListUserInviteRequest
	}

	// IDoInviteBuilder 邀请条目构造器
	IDoInviteBuilder interface {
		// ToAPI 转换为API对象
		ToAPI(*model.SysTeamInvite) *admin.InviteItem
		// ToAPIs 转换为API对象列表
		ToAPIs([]*model.SysTeamInvite) []*admin.InviteItem
	}

	doInviteBuilder struct {
		ctx      context.Context
		TeamInfo *bo.InviteTeamInfoParams
	}
)

func (d *doInviteBuilder) ToAPI(invite *model.SysTeamInvite) *admin.InviteItem {
	if types.IsNil(d) || types.IsNil(invite) {
		return nil
	}

	resItem := &admin.InviteItem{
		Id:         invite.ID,
		InviteType: api.InviteType(invite.InviteType),
	}
	teamInfo := d.TeamInfo
	if !types.IsNil(teamInfo) && !types.IsNil(teamInfo.TeamMap) {
		team := teamInfo.TeamMap[invite.TeamID]
		resItem.Team = NewParamsBuild(d.ctx).TeamModuleBuilder().DoTeamBuilder().ToAPI(team)
	}

	if !types.IsNil(teamInfo) && !types.IsNil(teamInfo.TeamRoles) {
		resItem.Roles = NewParamsBuild(d.ctx).RoleModuleBuilder().DoRoleBuilder().ToAPIs(teamInfo.TeamRoles)
	}
	return resItem
}

func (d *doInviteBuilder) ToAPIs(invites []*model.SysTeamInvite) []*admin.InviteItem {
	if types.IsNil(d) || types.IsNil(invites) {
		return nil
	}
	return types.SliceTo(invites, func(invite *model.SysTeamInvite) *admin.InviteItem {
		return d.ToAPI(invite)
	})
}

func (i *inviteModuleBuilder) DoInviteBuilder(teamInfo *bo.InviteTeamInfoParams) IDoInviteBuilder {
	return &doInviteBuilder{
		ctx:      i.ctx,
		TeamInfo: teamInfo,
	}
}

func (i *listInviteUserRequestBuilder) ToBo() *bo.QueryInviteListParams {
	if types.IsNil(i) || types.IsNil(i.ListUserInviteRequest) {
		return nil
	}
	return &bo.QueryInviteListParams{
		InviteType: vobj.InviteType(i.GetType()),
		Page:       types.NewPagination(i.GetPagination()),
	}
}

func (i updateInviteStatusRequestBuilder) ToBo() *bo.UpdateInviteStatusParams {
	if types.IsNil(i) || types.IsNil(i.UpdateInviteStatusRequest) {
		return nil
	}
	return &bo.UpdateInviteStatusParams{
		InviteID:   i.GetId(),
		InviteType: vobj.InviteType(i.GetType()),
	}
}

func (i *createInviteUserRequestBuilder) ToBo() *bo.InviteUserParams {
	if types.IsNil(i) || types.IsNil(i.InviteUserRequest) {
		return nil
	}
	return &bo.InviteUserParams{
		UserID:      0,
		TeamRoleIds: types.NewUint32SlicePointer(i.GetRoleIds()),
		InviteCode:  i.GetInviteCode(),
		TeamID:      middleware.GetTeamID(i.ctx),
		Role:        vobj.Role(i.GetRole()),
	}
}

func (i *inviteModuleBuilder) WithInviteUserRequest(request *inviteapi.InviteUserRequest) ICreateInviteUserRequestBuilder {
	return &createInviteUserRequestBuilder{
		ctx:               i.ctx,
		InviteUserRequest: request,
	}
}

func (i *inviteModuleBuilder) WithUpdateInviteStatusRequest(request *inviteapi.UpdateInviteStatusRequest) IUpdateInviteStatusRequestBuilder {
	return &updateInviteStatusRequestBuilder{
		ctx:                       i.ctx,
		UpdateInviteStatusRequest: request,
	}
}

func (i *inviteModuleBuilder) WithListInviteUserRequest(request *inviteapi.ListUserInviteRequest) IListInviteUserRequestBuilder {
	return &listInviteUserRequestBuilder{
		ctx:                   i.ctx,
		ListUserInviteRequest: request,
	}
}
