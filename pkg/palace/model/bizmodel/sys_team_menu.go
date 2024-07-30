package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysMenu = "sys_team_menus"

// SysTeamMenu mapped from table <sys_menus>
type SysTeamMenu struct {
	model.AllFieldModel
	Name     string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:菜单名称" json:"name"`
	EnName   string      `gorm:"column:en_name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:菜单英文名称" json:"en_name"`
	Path     string      `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__sa__path,priority:1;comment:api路径" json:"path"`
	Status   vobj.Status `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`
	Icon     string      `gorm:"column:icon;type:varchar(255);not null;comment:图标" json:"icon"`
	ParentID uint32      `gorm:"column:parent_id;type:int unsigned;not null;default:0;comment:父级ID" json:"parent_id"`
	Level    int32       `gorm:"column:level;type:int;not null;comment:层级" json:"level"`

	Parent *SysTeamMenu `gorm:"foreignKey:ParentID;references:ID" json:"parent"`
}

// String json string
func (c *SysTeamMenu) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamMenu) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamMenu) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysAPI's table name
func (*SysTeamMenu) TableName() string {
	return tableNameSysMenu
}
