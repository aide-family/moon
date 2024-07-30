package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeamMember = "sys_team_members"

// SysTeamMember mapped from table <sys_team_members>
type SysTeamMember struct {
	model.AllFieldModel
	UserID    uint32         `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id,priority:1;comment:系统用户ID" json:"user_id"` // 系统用户ID
	TeamID    uint32         `gorm:"column:team_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id,priority:2;comment:团队ID" json:"team_id"`   // 团队ID
	Status    vobj.Status    `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                              // 状态
	Role      vobj.Role      `gorm:"column:role;type:int;not null;comment:是否是管理员" json:"role"`                                                              // 是否是管理员
	TeamRoles []*SysTeamRole `gorm:"many2many:sys_team_member_roles" json:"team_roles"`
}

// String json string
func (c *SysTeamMember) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamMember) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamMember) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysTeamMember's table name
func (*SysTeamMember) TableName() string {
	return tableNameSysTeamMember
}
