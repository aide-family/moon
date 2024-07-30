package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeam = "sys_teams"

// SysTeam mapped from table <sys_teams>
type SysTeam struct {
	AllFieldModel
	Name     string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__name,priority:1;comment:团队空间名" json:"name"`               // 团队空间名
	Status   vobj.Status `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                       // 状态
	Remark   string      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                              // 备注
	Logo     string      `gorm:"column:logo;type:varchar(255);not null;comment:团队logo" json:"logo"`                                              // 团队logo
	LeaderID uint32      `gorm:"column:leader_id;type:int unsigned;not null;index:sys_teams__sys_users,priority:1;comment:负责人" json:"leader_id"` // 负责人
	UUID     string      `gorm:"column:uuid;type:varchar(64);not null" json:"uuid"`
}

// String json string
func (c *SysTeam) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeam) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeam) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysTeam's table name
func (*SysTeam) TableName() string {
	return tableNameSysTeam
}
