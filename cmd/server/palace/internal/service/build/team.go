package build

import (
	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/admin"
	"github.com/aide-cloud/moon/pkg/helper/model"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
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
func (b *TeamBuild) ToApi() *admin.Team {
	if types.IsNil(b) || types.IsNil(b.SysTeam) {
		return nil
	}
	return &admin.Team{
		Id:        b.ID,
		Name:      b.Name,
		Status:    api.Status(b.Status),
		Remark:    b.Remark,
		CreatedAt: b.CreatedAt.String(),
		UpdatedAt: b.UpdatedAt.String(),
		// TODO 从全局中取
		//Leader:    NewUserBuild(b.Leader).ToApi(),
		//Creator:   NewUserBuild(b.Creator).ToApi(),
		Logo:  b.Logo,
		Admin: nil,
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
