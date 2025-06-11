package system

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

var _ do.TeamRole = (*TeamRole)(nil)

const tableNameTeamRole = "team_roles"

type TeamRole struct {
	do.TeamModel
	Name    string            `gorm:"column:name;type:varchar(64);not null;comment:角色名" json:"name"`
	Remark  string            `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status  vobj.GlobalStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Members []*TeamMember     `gorm:"many2many:team_member_roles" json:"members"`
	Menus   []*Menu           `gorm:"many2many:team_role_menus" json:"menus"`
}

func (u *TeamRole) GetName() string {
	if u == nil {
		return ""
	}
	return u.Name
}

func (u *TeamRole) GetRemark() string {
	if u == nil {
		return ""
	}
	return u.Remark
}

func (u *TeamRole) GetStatus() vobj.GlobalStatus {
	if u == nil {
		return vobj.GlobalStatusUnknown
	}
	return u.Status
}

func (u *TeamRole) GetMembers() []do.TeamMember {
	if u == nil {
		return nil
	}
	return slices.Map(u.Members, func(m *TeamMember) do.TeamMember { return m })
}

func (u *TeamRole) GetMenus() []do.Menu {
	if u == nil {
		return nil
	}
	return slices.Map(u.Menus, func(m *Menu) do.Menu { return m })
}

func (u *TeamRole) TableName() string {
	return tableNameTeamRole
}
