package system

import (
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
)

var _ do.Menu = (*Menu)(nil)

const tableNameMenu = "sys_menus"

type Menu struct {
	do.BaseModel
	Name          string               `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__menu__name,priority:1;comment:菜单名称" json:"name"`
	MenuPath      string               `gorm:"column:menu_path;type:varchar(255);not null;uniqueIndex:idx__menu__menu_path,priority:1;comment:菜单路径" json:"menuPath"`
	MenuIcon      string               `gorm:"column:menu_icon;type:varchar(64);not null;comment:菜单图标" json:"menuIcon"`
	MenuType      vobj.MenuType        `gorm:"column:menu_type;type:tinyint(2);not null;comment:菜单系统类型" json:"menuType"`
	MenuCategory  vobj.MenuCategory    `gorm:"column:menu_category;type:tinyint(2);not null;comment:菜单类型" json:"menuCategory"`
	ApiPath       string               `gorm:"column:api_path;type:varchar(255);not null;comment:接口路径" json:"apiPath"`
	Status        vobj.GlobalStatus    `gorm:"column:status;type:tinyint(2);not null;comment:状态" json:"status"`
	ProcessType   vobj.MenuProcessType `gorm:"column:process_type;type:tinyint(2);not null;comment:处理类型" json:"processType"`
	ParentID      uint32               `gorm:"column:parent_id;type:int unsigned;not null;default:0;comment:父级id" json:"parentID"`
	Parent        *Menu                `gorm:"foreignKey:ParentID;references:ID" json:"parent"`
	RelyOnBrother bool                 `gorm:"column:rely_on_brother;type:tinyint(1);not null;default:0;comment:是否依赖兄弟节点" json:"relyOnBrother"`
}

func (u *Menu) GetName() string {
	return u.Name
}

func (u *Menu) GetMenuPath() string {
	return u.MenuPath
}

func (u *Menu) GetMenuIcon() string {
	return u.MenuIcon
}

func (u *Menu) GetMenuType() vobj.MenuType {
	return u.MenuType
}

func (u *Menu) GetMenuCategory() vobj.MenuCategory {
	return u.MenuCategory
}

func (u *Menu) GetApiPath() string {
	return u.ApiPath
}

func (u *Menu) GetStatus() vobj.GlobalStatus {
	return u.Status
}

func (u *Menu) GetProcessType() vobj.MenuProcessType {
	return u.ProcessType
}

func (u *Menu) GetParentID() uint32 {
	return u.ParentID
}

func (u *Menu) GetParent() do.Menu {
	return u.Parent
}

func (u *Menu) IsRelyOnBrother() bool {
	return u.RelyOnBrother
}

func (u *Menu) TableName() string {
	return tableNameMenu
}
