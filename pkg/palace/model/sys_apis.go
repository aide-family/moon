package model

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysAPI = "sys_apis"

// SysAPI mapped from table <sys_apis>
type SysAPI struct {
	AllFieldModel
	Name   string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__sa__name,priority:1;comment:api名称" json:"name"`  // api名称
	Path   string      `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__sa__path,priority:1;comment:api路径" json:"path"` // api路径
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`                                          // 状态
	Remark string      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                     // 备注
	Module int32       `gorm:"column:module;type:int;not null;comment:模块" json:"module"`                                              // 模块
	Domain int32       `gorm:"column:domain;type:int;not null;comment:领域" json:"domain"`                                              // 领域
}

// String json string
func (c *SysAPI) String() string {
	bs, _ := json.Marshal(c)
	return string(bs)
}

// UnmarshalBinary 实现redis存储
func (c *SysAPI) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, c)
}

// MarshalBinary 实现redis存储
func (c *SysAPI) MarshalBinary() (data []byte, err error) {
	return json.Marshal(c)
}

// TableName SysAPI's table name
func (*SysAPI) TableName() string {
	return tableNameSysAPI
}
