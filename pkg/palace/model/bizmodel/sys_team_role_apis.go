package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
)

const tableNameSysTeamRoleAPI = "sys_team_role_apis"

// SysTeamRoleAPI mapped from table <sys_team_role_apis>
type SysTeamRoleAPI struct {
	model.BaseModel
	SysTeamRoleID uint32 `gorm:"column:sys_team_role_id;type:int unsigned;primaryKey" json:"sys_team_role_id"`
	SysTeamAPIID  uint32 `gorm:"column:sys_api_id;type:int unsigned;primaryKey" json:"sys_team_api_id"`
}

// String json string
func (c *SysTeamRoleAPI) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// TableName SysTeamRoleAPI's table name
func (*SysTeamRoleAPI) TableName() string {
	return tableNameSysTeamRoleAPI
}
