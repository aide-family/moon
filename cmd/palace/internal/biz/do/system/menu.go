package system

import (
	"encoding/json"

	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/vobj"
	"gorm.io/plugin/soft_delete"
)

var _ do.Menu = (*Menu)(nil)

const tableNameMenu = "sys_menus"

type Menu struct {
	do.BaseModel
	DeletedAt     soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;default:0;uniqueIndex:idx__menu__name__deleted_at,priority:2;" json:"deleted_at,omitempty"`
	Name          string                `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__menu__name__deleted_at,priority:1;comment:menu name" json:"name"`
	MenuPath      string                `gorm:"column:menu_path;type:varchar(255);not null;default:'';comment:menu path" json:"menuPath"`
	MenuIcon      string                `gorm:"column:menu_icon;type:varchar(64);not null;default:'';comment:menu icon" json:"menuIcon"`
	MenuType      vobj.MenuType         `gorm:"column:menu_type;type:tinyint(2);not null;default:0;comment:menu system type" json:"menuType"`
	MenuCategory  vobj.MenuCategory     `gorm:"column:menu_category;type:tinyint(2);not null;default:0;comment:menu category" json:"menuCategory"`
	APIPath       string                `gorm:"column:api_path;type:varchar(255);not null;default:'';comment:API path" json:"apiPath"`
	Status        vobj.GlobalStatus     `gorm:"column:status;type:tinyint(2);not null;default:0;comment:status" json:"status"`
	ProcessType   vobj.MenuProcessType  `gorm:"column:process_type;type:tinyint(2);not null;default:0;comment:process type" json:"processType"`
	ParentID      uint32                `gorm:"column:parent_id;type:int unsigned;not null;default:0;comment:parent ID" json:"parentID"`
	Parent        *Menu                 `gorm:"foreignKey:ParentID;references:ID" json:"parent"`
	RelyOnBrother bool                  `gorm:"column:rely_on_brother;type:tinyint(1);not null;default:0;comment:whether to rely on sibling node" json:"relyOnBrother"`
	Sort          uint32                `gorm:"column:sort;type:int unsigned;not null;default:0;comment:sort" json:"sort"`
}

func (u *Menu) GetDeletedAt() soft_delete.DeletedAt {
	return u.DeletedAt
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

func (u *Menu) GetAPIPath() string {
	return u.APIPath
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

func (u *Menu) GetSort() uint32 {
	return u.Sort
}

func (u *Menu) MarshalBinary() ([]byte, error) {
	return json.Marshal(u)
}

func (u *Menu) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, u)
}

func (u *Menu) UniqueKey() string {
	return u.APIPath
}

func (u *Menu) TableName() string {
	return tableNameMenu
}
