package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/plugin/soft_delete"
)

const TableNameSysMenu = "sys_team_menus"

// SysTeamMenu mapped from table <sys_menus>
type SysTeamMenu struct {
	ID        uint32                `gorm:"column:id;type:int unsigned;primaryKey;autoIncrement:true" json:"id"`
	CreatedAt types.Time            `gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:创建时间" json:"created_at"` // 创建时间
	UpdatedAt types.Time            `gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP;comment:更新时间" json:"updated_at"` // 更新时间
	DeletedAt soft_delete.DeletedAt `gorm:"column:deleted_at;type:bigint;not null;comment:删除时间" json:"deleted_at"`
	Name      string                `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:菜单名称" json:"name"`
	EnName    string                `gorm:"column:en_name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:菜单英文名称" json:"en_name"`
	Path      string                `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__sa__path,priority:1;comment:api路径" json:"path"`
	Status    vobj.Status           `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`
	Icon      string                `gorm:"column:icon;type:varchar(255);not null;comment:图标" json:"icon"`
	ParentID  uint32                `gorm:"column:parent_id;type:int unsigned;not null;default:0;comment:父级ID" json:"parent_id"`
	Level     int32                 `gorm:"column:level;type:int;not null;comment:层级" json:"level"`

	Parent *SysTeamMenu `gorm:"foreignKey:ParentID;references:ID" json:"parent"`
}

// String json string
func (c *SysTeamMenu) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *SysTeamMenu) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *SysTeamMenu) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysAPI's table name
func (*SysTeamMenu) TableName() string {
	return TableNameSysMenu
}
