package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
)

const tableNameSysTeamRoleAPI = "sys_team_role_apis"

// SysTeamRoleAPI mapped from table <sys_team_role_apis>
type SysTeamRoleAPI struct {
	SysTeamRoleID uint32 `gorm:"primaryKey;column:sys_team_role_id;type:int unsigned;primaryKey" json:"sys_team_role_id"`
	SysTeamAPIID  uint32 `gorm:"primaryKey;column:sys_api_id;type:int unsigned;primaryKey" json:"sys_team_api_id"`
}

// String json string
func (c *SysTeamRoleAPI) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// TableName SysTeamRoleAPI's table name
func (*SysTeamRoleAPI) TableName() string {
	return tableNameSysTeamRoleAPI
}
