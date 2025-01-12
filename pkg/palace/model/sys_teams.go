package model

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeam = "sys_teams"

// SysTeam mapped from table <sys_teams>
type SysTeam struct {
	AllFieldModel
	Name     string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__team__name,priority:1;comment:团队空间名" json:"name"`         // 团队空间名
	Status   vobj.Status `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                       // 状态
	Remark   string      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                              // 备注
	Logo     string      `gorm:"column:logo;type:varchar(255);not null;comment:团队logo" json:"logo"`                                              // 团队logo
	LeaderID uint32      `gorm:"column:leader_id;type:int unsigned;not null;index:sys_teams__sys_users,priority:1;comment:负责人" json:"leader_id"` // 负责人
	UUID     string      `gorm:"column:uuid;type:varchar(64);not null" json:"uuid"`

	Admins []uint32 `gorm:"-" json:"admins"`
}

// String json string
func (c *SysTeam) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeam) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeam) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeam's table name
func (*SysTeam) TableName() string {
	return tableNameSysTeam
}

// GetName get name
func (c *SysTeam) GetName() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Name
}

// GetStatus get status
func (c *SysTeam) GetStatus() vobj.Status {
	if types.IsNil(c) {
		return vobj.Status(0)
	}
	return c.Status
}

// GetRemark get remark
func (c *SysTeam) GetRemark() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Remark
}

// GetLogo get logo
func (c *SysTeam) GetLogo() string {
	if types.IsNil(c) {
		return ""
	}
	return c.Logo
}

// GetLeaderID get leader id
func (c *SysTeam) GetLeaderID() uint32 {
	if types.IsNil(c) {
		return 0
	}
	return c.LeaderID
}

// GetUUID get uuid
func (c *SysTeam) GetUUID() string {
	if types.IsNil(c) {
		return ""
	}
	return c.UUID
}
