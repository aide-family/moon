package system

import (
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

var _ do.Role = (*Role)(nil)

const tableNameRole = "sys_roles"

type Role struct {
	do.CreatorModel
	Name   string            `gorm:"column:name;type:varchar(64);not null;comment:角色名" json:"name"`
	Remark string            `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status vobj.GlobalStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Users  []*User           `gorm:"many2many:sys_user_roles" json:"users"`
	Menus  []*Menu           `gorm:"many2many:sys_role_menus" json:"menus"`
}

func (u *Role) GetName() string {
	if u == nil {
		return ""
	}
	return u.Name
}

func (u *Role) GetRemark() string {
	if u == nil {
		return ""
	}
	return u.Remark
}

func (u *Role) GetStatus() vobj.GlobalStatus {
	if u == nil {
		return vobj.GlobalStatusUnknown
	}
	return u.Status
}

func (u *Role) GetUsers() []do.User {
	if u == nil {
		return nil
	}
	return slices.Map(u.Users, func(v *User) do.User { return v })
}

func (u *Role) GetMenus() []do.Menu {
	if u == nil {
		return nil
	}
	return slices.Map(u.Menus, func(v *Menu) do.Menu { return v })
}

func (u *Role) TableName() string {
	return tableNameRole
}
