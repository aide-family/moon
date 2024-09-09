package builder

import (
	"context"

	"github.com/aide-family/moon/api"
	adminapi "github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

var _ ITeamMemberModuleBuilder = (*teamMemberModuleBuilder)(nil)

type (
	teamMemberModuleBuilder struct {
		ctx context.Context
	}

	ITeamMemberModuleBuilder interface {
		DoTeamMemberBuilder() IDoTeamMemberBuilder
	}

	IDoTeamMemberBuilder interface {
		ToAPI(*bizmodel.SysTeamMember) *adminapi.TeamMemberItem
		ToAPIs([]*bizmodel.SysTeamMember) []*adminapi.TeamMemberItem
		ToSelect(*bizmodel.SysTeamMember) *adminapi.SelectItem
		ToSelects([]*bizmodel.SysTeamMember) []*adminapi.SelectItem
	}

	doTeamMemberBuilder struct {
		ctx context.Context
	}
)

func (d *doTeamMemberBuilder) ToAPI(member *bizmodel.SysTeamMember) *adminapi.TeamMemberItem {
	if types.IsNil(d) || types.IsNil(member) {
		return nil
	}

	return &adminapi.TeamMemberItem{
		UserId:    member.UserID,
		Id:        member.ID,
		Role:      api.Role(member.Role),
		Status:    api.Status(member.Status),
		CreatedAt: member.CreatedAt.String(),
		UpdatedAt: member.UpdatedAt.String(),
		User:      nil, // TODO user
	}
}

func (d *doTeamMemberBuilder) ToAPIs(members []*bizmodel.SysTeamMember) []*adminapi.TeamMemberItem {
	if types.IsNil(d) || types.IsNil(members) {
		return nil
	}

	return types.SliceTo(members, func(member *bizmodel.SysTeamMember) *adminapi.TeamMemberItem {
		return d.ToAPI(member)
	})
}

func (d *doTeamMemberBuilder) ToSelect(member *bizmodel.SysTeamMember) *adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(member) {
		return nil
	}
	userInfo := new(model.SysUser) // TODO 获取实际的user
	return &adminapi.SelectItem{
		Value:    member.ID,
		Label:    userInfo.Username,
		Children: nil,
		Disabled: member.DeletedAt > 0 || !member.Status.IsEnable() || userInfo.DeletedAt > 0 || !userInfo.Status.IsEnable(),
		Extend: &adminapi.SelectExtend{
			Remark: userInfo.Remark,
			Image:  userInfo.Avatar,
		},
	}
}

func (d *doTeamMemberBuilder) ToSelects(members []*bizmodel.SysTeamMember) []*adminapi.SelectItem {
	if types.IsNil(d) || types.IsNil(members) {
		return nil
	}

	return types.SliceTo(members, func(member *bizmodel.SysTeamMember) *adminapi.SelectItem {
		return d.ToSelect(member)
	})
}

func (t *teamMemberModuleBuilder) DoTeamMemberBuilder() IDoTeamMemberBuilder {
	return &doTeamMemberBuilder{ctx: t.ctx}
}
