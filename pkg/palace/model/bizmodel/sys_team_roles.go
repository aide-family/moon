package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameSysTeamRole = "sys_team_roles"

// SysTeamRole mapped from table <sys_team_roles>
type SysTeamRole struct {
	model.AllFieldModel
	TeamID uint32        `gorm:"column:team_id;type:int unsigned;not null;comment:团队ID" json:"team_id"` // 团队ID
	Name   string        `gorm:"column:name;type:varchar(64);not null;comment:角色名称" json:"name"`        // 角色名称
	Status int           `gorm:"column:status;type:int;not null;comment:状态" json:"status"`              // 状态
	Remark string        `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`     // 备注
	Apis   []*SysTeamAPI `gorm:"many2many:sys_team_role_apis" json:"apis"`
}

// String json string
func (c *SysTeamRole) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamRole) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamRole) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysTeamRole's table name
func (*SysTeamRole) TableName() string {
	return tableNameSysTeamRole
}
