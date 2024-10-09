package builder

import (
	"context"

	"github.com/aide-family/moon/api/admin"
	inviteapi "github.com/aide-family/moon/api/admin/invite"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

var _ InviteModuleBuilder = (*inviteModuleBuilder)(nil)

type (
	InviteModuleBuilder interface {
		WithInviteUserRequest(*inviteapi.InviteUserRequest) ICreateInviteUserRequestBuilder
		WithUpdateInviteStatusRequest(*inviteapi.UpdateInviteStatusRequest) IUpdateInviteStatusRequestBuilder
		WithListInviteUserRequest(*inviteapi.ListInviteRequest) IListInviteUserRequestBuilder
		DoInviteBuilder() IDoInviteBuilder
	}
	inviteModuleBuilder struct {
		ctx context.Context
	}

	ICreateInviteUserRequestBuilder interface {
		ToBo() *bo.InviteUserParams
	}

	createInviteUserRequestBuilder struct {
		ctx context.Context
		*inviteapi.InviteUserRequest
	}

	IUpdateInviteStatusRequestBuilder interface {
		ToBo() *bo.UpdateInviteStatusParams
	}
	updateInviteStatusRequestBuilder struct {
		ctx context.Context
		*inviteapi.UpdateInviteStatusRequest
	}

	IListInviteUserRequestBuilder interface {
		ToBo() *bo.QueryInviteListParams
	}

	listInviteUserRequestBuilder struct {
		ctx context.Context
		*inviteapi.ListInviteRequest
	}

	IDoInviteBuilder interface {
		ToAPI(*bizmodel.SysTeamInvite) *admin.InviteItem
		ToAPIs([]*bizmodel.SysTeamInvite) *admin.InviteItem
	}

	doInviteBuilder struct {
		ctx context.Context
	}
)

func (d *doInviteBuilder) ToAPI(invite *bizmodel.SysTeamInvite) *admin.InviteItem {
	//TODO implement me
	panic("implement me")
}

func (d *doInviteBuilder) ToAPIs(invites []*bizmodel.SysTeamInvite) *admin.InviteItem {
	//TODO implement me
	panic("implement me")
}

func (i *inviteModuleBuilder) DoInviteBuilder() IDoInviteBuilder {
	return &doInviteBuilder{
		ctx: i.ctx,
	}
}

func (i *listInviteUserRequestBuilder) ToBo() *bo.QueryInviteListParams {
	if types.IsNil(i) || types.IsNil(i.ListInviteRequest) {
		return nil
	}
	return &bo.QueryInviteListParams{
		InviteType: vobj.InviteType(i.GetType()),
		Keyword:    i.GetKeyword(),
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
		TeamRoleID: i.GetRoleId(),
		InviteCode: i.GetInviteCode(),
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

func (i *inviteModuleBuilder) WithListInviteUserRequest(request *inviteapi.ListInviteRequest) IListInviteUserRequestBuilder {
	return &listInviteUserRequestBuilder{
		ctx:               i.ctx,
		ListInviteRequest: request,
	}
}
