package system

import (
	"time"

	"gorm.io/plugin/soft_delete"

	"github.com/moon-monitor/moon/cmd/palace/internal/biz/do"
	"github.com/moon-monitor/moon/cmd/palace/internal/biz/vobj"
	"github.com/moon-monitor/moon/pkg/util/slices"
)

var _ do.Menu = (*Menu)(nil)

const tableNameMenu = "sys_menus"

type Menu struct {
	do.BaseModel
	Name      string            `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__menu__name,priority:1;comment:菜单名称" json:"name"`
	Path      string            `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__menu__path,priority:1;comment:菜单路径" json:"path"`
	Status    vobj.GlobalStatus `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	Icon      string            `gorm:"column:icon;type:varchar(64);not null;comment:图标" json:"icon"`
	ParentID  uint32            `gorm:"column:parent_id;type:int unsigned;not null;default:0;comment:父级id" json:"parentID"`
	Type      vobj.MenuType     `gorm:"column:type;type:tinyint(2);not null;comment:菜单类型" json:"type"`
	Parent    *Menu             `gorm:"foreignKey:ParentID;references:ID" json:"parent"`
	Resources []*Resource       `gorm:"many2many:sys_menu_resources" json:"resources"`
}

func (u *Menu) GetCreatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.CreatedAt
}

func (u *Menu) GetUpdatedAt() time.Time {
	if u == nil {
		return time.Time{}
	}
	return u.UpdatedAt
}

func (u *Menu) GetDeletedAt() soft_delete.DeletedAt {
	if u == nil {
		return 0
	}
	return u.DeletedAt
}

func (u *Menu) GetID() uint32 {
	if u == nil {
		return 0
	}
	return u.ID
}

func (u *Menu) GetName() string {
	if u == nil {
		return ""
	}
	return u.Name
}

func (u *Menu) GetPath() string {
	if u == nil {
		return ""
	}
	return u.Path
}

func (u *Menu) GetStatus() vobj.GlobalStatus {
	if u == nil {
		return vobj.GlobalStatusUnknown
	}
	return u.Status
}

func (u *Menu) GetIcon() string {
	if u == nil {
		return ""
	}
	return u.Icon
}

func (u *Menu) GetParentID() uint32 {
	if u == nil {
		return 0
	}
	return u.ParentID
}

func (u *Menu) GetType() vobj.MenuType {
	if u == nil {
		return vobj.MenuTypeUnknown
	}
	return u.Type
}

func (u *Menu) GetResources() []do.Resource {
	if u == nil {
		return nil
	}
	return slices.Map(u.Resources, func(r *Resource) do.Resource { return r })
}

func (u *Menu) GetParent() do.Menu {
	if u == nil {
		return nil
	}
	return u.Parent
}

func (u *Menu) TableName() string {
	return tableNameMenu
}
