package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	teamapi "github.com/aide-family/moon/api/admin/team"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type (
	TeamModelBuilder interface {
		ToApi() *admin.Team
	}

	TeamRequestBuilder interface {
		ToCreateTeamBO() *bo.CreateTeamParams

		ToUpdateRoleBO() *bo.UpdateTeamParams

		ToQueryTeamListBO() *bo.QueryTeamListParams

		ToListTeamMemberBO() *bo.ListTeamMemberParams

		ToTeamListBO() *bo.QueryTeamListParams

		ToAddTeamMemberBO() *bo.AddTeamMemberParams

		WithLeaderId(leaderId uint32)
	}

	teamBuilder struct {
		// model
		*model.SysTeam

		// request
		CreateRoleRequest     *teamapi.CreateTeamRequest
		UpdateTeamRequest     *teamapi.UpdateTeamRequest
		ListTeamRequest       *teamapi.ListTeamRequest
		ListTeamMemberRequest *teamapi.ListTeamMemberRequest
		AddTeamMemberRequest  *teamapi.AddTeamMemberRequest
		LeaderId              uint32

		// context
		ctx context.Context
	}

	TeamRoleBuilder interface {
		ToSelect() *admin.Select
		ToApi() *admin.TeamRole
	}

	teamRoleBuilder struct {
		*bizmodel.SysTeamRole
		ctx context.Context
	}
)

func (b *teamBuilder) ToAddTeamMemberBO() *bo.AddTeamMemberParams {
	return &bo.AddTeamMemberParams{
		ID: b.AddTeamMemberRequest.GetId(),
		Members: types.SliceTo(b.AddTeamMemberRequest.GetMembers(), func(member *teamapi.AddTeamMemberRequest_MemberItem) *bo.AddTeamMemberItem {
			return &bo.AddTeamMemberItem{
				UserID:  member.GetUserId(),
				Role:    vobj.Role(member.GetRole()),
				RoleIds: member.GetRoles(),
			}
		}),
	}
}

func (b *teamBuilder) ToQueryTeamListBO() *bo.QueryTeamListParams {
	//TODO implement me
	panic("implement me")
}

// ToApi 转换为API层数据
func (b *teamBuilder) ToApi() *admin.Team {
	if types.IsNil(b) || types.IsNil(b.SysTeam) {
		return nil
	}
	cache := runtimecache.GetRuntimeCache()
	return &admin.Team{
		Id:        b.ID,
		Name:      b.Name,
		Status:    api.Status(b.Status),
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		Leader:    NewBuilder().WithApiUserBo(cache.GetUser(b.ctx, b.LeaderID)).ToApi(),
		Creator:   NewBuilder().WithApiUserBo(cache.GetUser(b.ctx, b.CreatorID)).ToApi(),
		Logo:      b.Logo,
		// 从全局中取
		Admin: types.SliceTo(cache.GetTeamAdminList(b.ctx, b.ID), func(item *bizmodel.SysTeamMember) *admin.TeamMember {
			return NewBuilder().WithApiTeamMember(item).ToApi(b.ctx)
		}),
	}
}

func (b *teamBuilder) ToCreateTeamBO() *bo.CreateTeamParams {
	return &bo.CreateTeamParams{
		Name:     b.CreateRoleRequest.GetName(),
		Remark:   b.CreateRoleRequest.GetRemark(),
		Logo:     b.CreateRoleRequest.GetLogo(),
		Status:   vobj.Status(b.CreateRoleRequest.GetStatus()),
		LeaderID: b.LeaderID,
		Admins:   b.CreateRoleRequest.GetAdminIds(),
	}
}

func (b *teamBuilder) ToUpdateRoleBO() *bo.UpdateTeamParams {
	return &bo.UpdateTeamParams{
		ID:     b.UpdateTeamRequest.GetId(),
		Name:   b.UpdateTeamRequest.GetName(),
		Remark: b.UpdateTeamRequest.GetRemark(),
		Logo:   b.UpdateTeamRequest.GetLogo(),
		Status: vobj.Status(b.UpdateTeamRequest.GetStatus()),
	}
}

func (b *teamBuilder) ToQueryTeamList() *bo.QueryTeamListParams {
	return &bo.QueryTeamListParams{
		Page:      types.NewPagination(b.ListTeamRequest.GetPagination()),
		Keyword:   b.ListTeamRequest.GetKeyword(),
		Status:    vobj.Status(b.ListTeamRequest.GetStatus()),
		CreatorID: b.ListTeamRequest.GetCreatorId(),
		LeaderID:  b.ListTeamRequest.GetLeaderId(),
	}
}

func (b *teamBuilder) ToListTeamMemberBO() *bo.ListTeamMemberParams {
	return &bo.ListTeamMemberParams{
		Page:    types.NewPagination(b.ListTeamMemberRequest.GetPagination()),
		ID:      b.ListTeamMemberRequest.GetId(),
		Keyword: b.ListTeamMemberRequest.GetKeyword(),
		Role:    vobj.Role(b.ListTeamMemberRequest.GetRole()),
		Gender:  vobj.Gender(b.ListTeamMemberRequest.GetGender()),
		Status:  vobj.Status(b.ListTeamMemberRequest.GetStatus()),
	}
}

func (b *teamBuilder) ToTeamListBO() *bo.QueryTeamListParams {
	return &bo.QueryTeamListParams{
		Page:      types.NewPagination(b.ListTeamRequest.GetPagination()),
		Keyword:   b.ListTeamRequest.GetKeyword(),
		Status:    vobj.Status(b.ListTeamRequest.GetStatus()),
		CreatorID: b.ListTeamRequest.GetCreatorId(),
		LeaderID:  b.ListTeamRequest.GetLeaderId(),
	}
}

func (b *teamBuilder) WithLeaderId(leaderId uint32) {
	b.LeaderId = leaderId
}

func (b *teamRoleBuilder) ToApi() *admin.TeamRole {
	if types.IsNil(b) || types.IsNil(b.SysTeamRole) {
		return nil
	}
	return &admin.TeamRole{
		Id:        b.ID,
		Name:      b.Name,
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		Status:    api.Status(b.Status),
		Resources: types.SliceTo(b.Apis, func(item *bizmodel.SysTeamAPI) *admin.ResourceItem {
			return NewTeamResourceBuilder(item).ToApi()
		}),
	}
}

// ToSelect 转换为Select数据
func (b *teamRoleBuilder) ToSelect() *admin.Select {
	if types.IsNil(b) || types.IsNil(b.SysTeamRole) {
		return nil
	}
	return &admin.Select{
		Value:    b.ID,
		Label:    b.Name,
		Disabled: b.DeletedAt > 0 || !vobj.Status(b.Status).IsEnable(),
	}
}
