package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeamMember = "sys_team_members"

// SysTeamMember mapped from table <sys_team_members>
type SysTeamMember struct {
	AllFieldModel
	UserID    uint32         `gorm:"column:user_id;type:int unsigned;not null;uniqueIndex:idx__user_id__team__id,priority:1;comment:系统用户ID" json:"user_id"` // 系统用户ID
	Status    vobj.Status    `gorm:"column:status;type:int;not null;comment:状态" json:"status"`                                                              // 状态
	Role      vobj.Role      `gorm:"column:role;type:int;not null;comment:是否是管理员" json:"role"`                                                              // 是否是管理员
	TeamRoles []*SysTeamRole `gorm:"many2many:sys_team_member_roles" json:"team_roles"`
}

// GetUserID get user id
func (c *SysTeamMember) GetUserID() uint32 {
	if c == nil {
		return 0
	}
	return c.UserID
}

// String json string
func (c *SysTeamMember) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamMember) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamMember) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeamMember's table name
func (*SysTeamMember) TableName() string {
	return tableNameSysTeamMember
}

// GetTeamRoles 获取团队角色
func (c *SysTeamMember) GetTeamRoles() []*SysTeamRole {
	if c == nil {
		return nil
	}
	return c.TeamRoles
}

// GetStatus 获取状态
func (c *SysTeamMember) GetStatus() vobj.Status {
	if c == nil {
		return vobj.StatusUnknown
	}
	return c.Status
}

// GetRole 获取角色
func (c *SysTeamMember) GetRole() vobj.Role {
	if c == nil {
		return vobj.RoleAll
	}
	return c.Role
}
