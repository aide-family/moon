package model

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeamEmail = "sys_team_email"

// SysTeamEmail mapped from table <sys_team_email>
type SysTeamEmail struct {
	AllFieldModel
	TeamID uint32      `gorm:"column:team_id;type:int unsigned;not null;index:sys_teams__sys_team_email,unique,priority:1;comment:团队id" json:"team_id"`
	User   string      `gorm:"column:user;type:varchar(255);not null;comment:邮箱" json:"user"`
	Pass   string      `gorm:"column:pass;type:varchar(255);not null;comment:密码" json:"pass"`
	Host   string      `gorm:"column:host;type:varchar(255);not null;comment:主机" json:"host"`
	Port   uint32      `gorm:"column:port;type:int unsigned;not null;comment:端口" json:"port"`
	Remark string      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;default:1;comment:启用状态1:启用;2禁用" json:"status"`
}

// String json string
func (c *SysTeamEmail) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamEmail) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamEmail) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeamEmail's table name
func (*SysTeamEmail) TableName() string {
	return tableNameSysTeamEmail
}
