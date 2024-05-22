package build

import (
	"github.com/aide-cloud/moon/api"
	"github.com/aide-cloud/moon/api/admin"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/do/model"
	"github.com/aide-cloud/moon/pkg/types"
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
		Leader:    NewUserBuild(b.Leader).ToApi(),
		Creator:   NewUserBuild(b.Creator).ToApi(),
		Logo:      b.Logo,
		Admin:     nil,
	}
}
