package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

const tableNameSysTeamMemberRole = "sys_team_member_roles"

// SysTeamMemberRole mapped from table <sys_team_member_roles>
type SysTeamMemberRole struct {
	SysTeamMemberID uint32 `gorm:"primaryKey;column:sys_team_member_id;type:int unsigned;primaryKey;uniqueIndex:idx__user_id__team_id__role_id,priority:1;comment:团队用户ID" json:"sys_team_member_id"` // 团队用户ID
	SysTeamRoleID   uint32 `gorm:"primaryKey;column:sys_team_role_id;type:int unsigned;primaryKey;uniqueIndex:idx__user_id__team_id__role_id,priority:2;comment:团队角色ID" json:"sys_team_role_id"`     // 团队角色ID
}

// String json string
func (c *SysTeamMemberRole) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// TableName SysTeamMemberRole's table name
func (*SysTeamMemberRole) TableName() string {
	return tableNameSysTeamMemberRole
}
