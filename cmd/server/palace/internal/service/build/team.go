package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type TeamBuilder struct {
	*model.SysTeam
}

func NewTeamBuilder(team *model.SysTeam) *TeamBuilder {
	return &TeamBuilder{
		SysTeam: team,
	}
}

// ToApi 转换为API层数据
func (b *TeamBuilder) ToApi(ctx context.Context) *admin.Team {
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
		Leader:    NewUserBuilder(cache.GetUser(ctx, b.LeaderID)).ToApi(),
		Creator:   NewUserBuilder(cache.GetUser(ctx, b.CreatorID)).ToApi(),
		Logo:      b.Logo,
		// 从全局中取
		Admin: types.SliceTo(cache.GetTeamAdminList(ctx, b.ID), func(item *bizmodel.SysTeamMember) *admin.TeamMember {
			return NewTeamMemberBuilder(item).ToApi(ctx)
		}),
	}
}

type TeamRoleBuilder struct {
	*bizmodel.SysTeamRole
}

func NewTeamRoleBuilder(role *bizmodel.SysTeamRole) *TeamRoleBuilder {
	return &TeamRoleBuilder{
		SysTeamRole: role,
	}
}

func (b *TeamRoleBuilder) ToApi() *admin.TeamRole {
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
func (b *TeamRoleBuilder) ToSelect() *admin.Select {
	if types.IsNil(b) || types.IsNil(b.SysTeamRole) {
		return nil
	}
	return &admin.Select{
		Value:    b.ID,
		Label:    b.Name,
		Disabled: b.DeletedAt > 0 || !vobj.Status(b.Status).IsEnable(),
	}
}
