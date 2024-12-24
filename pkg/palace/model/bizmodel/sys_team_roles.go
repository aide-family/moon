package bizmodel

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeamRole = "sys_team_roles"

// SysTeamRole mapped from table <sys_team_roles>
type SysTeamRole struct {
	AllFieldModel
	Name    string           `gorm:"column:name;type:varchar(64);not null;comment:角色名称" json:"name"`    // 角色名称
	Status  vobj.Status      `gorm:"column:status;type:int;not null;comment:状态" json:"status"`          // 状态
	Remark  string           `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"` // 备注
	Apis    []*SysTeamAPI    `gorm:"many2many:sys_team_role_apis" json:"apis"`
	Members []*SysTeamMember `gorm:"many2many:sys_team_member_roles" json:"members"`
}

// String json string
func (c *SysTeamRole) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamRole) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamRole) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeamRole's table name
func (*SysTeamRole) TableName() string {
	return tableNameSysTeamRole
}
