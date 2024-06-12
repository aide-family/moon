package build

import (
	"context"

	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/api/admin"
	"github.com/aide-family/moon/cmd/server/palace/internal/data/runtimecache"
	"github.com/aide-family/moon/pkg/helper/model"
	"github.com/aide-family/moon/pkg/helper/model/bizmodel"
	"github.com/aide-family/moon/pkg/types"
	"github.com/aide-family/moon/pkg/vobj"
)

type TeamBuild struct {
	*model.SysTeam
}

func NewTeamBuild(team *model.SysTeam) *TeamBuild {
	return &TeamBuild{
		SysTeam: team,
	}
}

// ToApi 转换为API层数据
func (b *TeamBuild) ToApi(ctx context.Context) *admin.Team {
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
		Leader:    NewUserBuild(cache.GetUser(ctx, b.LeaderID)).ToApi(),
		Creator:   NewUserBuild(cache.GetUser(ctx, b.CreatorID)).ToApi(),
		Logo:      b.Logo,
		// 从全局中取
		Admin: types.SliceTo(cache.GetTeamAdminList(ctx, b.ID), func(item *bizmodel.SysTeamMember) *admin.TeamMember {
			return NewTeamMemberBuild(item).ToApi(ctx)
		}),
	}
}

type TeamRoleBuild struct {
	*bizmodel.SysTeamRole
}

func NewTeamRoleBuild(role *bizmodel.SysTeamRole) *TeamRoleBuild {
	return &TeamRoleBuild{
		SysTeamRole: role,
	}
}

func (b *TeamRoleBuild) ToApi() *admin.TeamRole {
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
			return NewTeamResourceBuild(item).ToApi()
		}),
	}
}

// ToSelect 转换为Select数据
func (b *TeamRoleBuild) ToSelect() *admin.Select {
	if types.IsNil(b) || types.IsNil(b.SysTeamRole) {
		return nil
	}
	return &admin.Select{
		Value:    b.ID,
		Label:    b.Name,
		Disabled: b.DeletedAt > 0 || !vobj.Status(b.Status).IsEnable(),
	}
}
