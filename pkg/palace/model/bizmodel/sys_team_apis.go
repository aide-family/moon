package bizmodel

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysAPI = "sys_team_apis"

// SysTeamAPI mapped from table <sys_apis>
type SysTeamAPI struct {
	model.AllFieldModel
	Name   string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:api名称" json:"name"`  // api名称
	Path   string      `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__sa__path,priority:1;comment:api路径" json:"path"` // api路径
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`                                          // 状态
	Remark string      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                     // 备注
	Module int32       `gorm:"column:module;type:int;not null;comment:模块" json:"module"`                                              // 模块
	Domain int32       `gorm:"column:domain;type:int;not null;comment:领域" json:"domain"`                                              // 领域
}

// String json string
func (c *SysTeamAPI) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis 存储实现
func (c *SysTeamAPI) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary redis 存储实现
func (c *SysTeamAPI) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysTeamAPI's table name
func (*SysTeamAPI) TableName() string {
	return tableNameSysAPI
}
