package bizmodel

import (
	"context"
	"encoding/json"

	"gorm.io/gen"
	"gorm.io/gorm"
)

const TableNameSysTeamRoleAPI = "sys_team_role_apis"

// SysTeamRoleAPI mapped from table <sys_team_role_apis>
type SysTeamRoleAPI struct {
	SysTeamRoleID uint32 `gorm:"column:sys_team_role_id;type:int unsigned;primaryKey" json:"sys_team_role_id"`
	SysTeamAPIID  uint32 `gorm:"column:sys_api_id;type:int unsigned;primaryKey" json:"sys_team_api_id"`
}

// String json string
func (c *SysTeamRoleAPI) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

func (c *SysTeamRoleAPI) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

func (c *SysTeamRoleAPI) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// Create func
func (c *SysTeamRoleAPI) Create(ctx context.Context, tx *gorm.DB) error {
	return tx.WithContext(ctx).Create(c).Error
}

// Update func
func (c *SysTeamRoleAPI) Update(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Model(c).Where(conds).Updates(c).Error
}

// Delete func
func (c *SysTeamRoleAPI) Delete(ctx context.Context, tx *gorm.DB, conds []gen.Condition) error {
	return tx.WithContext(ctx).Where(conds).Delete(c).Error
}

// TableName SysTeamRoleAPI's table name
func (*SysTeamRoleAPI) TableName() string {
	return TableNameSysTeamRoleAPI
}
