package bizmodel

import (
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

const tableNameSysTeamAPI = "sys_team_apis"

var _ imodel.IResource = (*SysTeamAPI)(nil)

// SysTeamAPI API资源管理
type SysTeamAPI struct {
	AllFieldModel
	Name   string      `gorm:"column:name;type:varchar(64);not null;uniqueIndex:idx__team_api__name,priority:1;comment:api名称" json:"name"`  // api名称
	Path   string      `gorm:"column:path;type:varchar(255);not null;uniqueIndex:idx__team_api__path,priority:1;comment:api路径" json:"path"` // api路径
	Status vobj.Status `gorm:"column:status;type:tinyint;not null;comment:状态" json:"status"`                                                // 状态
	Remark string      `gorm:"column:remark;type:varchar(255);not null;comment:备注" json:"remark"`                                           // 备注
	Module int32       `gorm:"column:module;type:int;not null;comment:模块" json:"module"`                                                    // 模块
	Domain int32       `gorm:"column:domain;type:int;not null;comment:领域" json:"domain"`                                                    // 领域
	Allow  vobj.Allow  `gorm:"column:allow;type:tinyint;not null;comment:允许类型" json:"allow"`                                                // 放行规则
}

// GetAllow 获取放行规则
func (c *SysTeamAPI) GetAllow() vobj.Allow {
	if c == nil {
		return 0
	}

	return c.Allow
}

// GetName 获取API名称
func (c *SysTeamAPI) GetName() string {
	if c == nil {
		return ""
	}

	return c.Name
}

// GetPath 获取API路径
func (c *SysTeamAPI) GetPath() string {
	if c == nil {
		return ""
	}

	return c.Path
}

// GetStatus 获取状态
func (c *SysTeamAPI) GetStatus() vobj.Status {
	if c == nil {
		return 0
	}
	return c.Status
}

// GetRemark 获取备注
func (c *SysTeamAPI) GetRemark() string {
	if c == nil {
		return ""
	}

	return c.Remark
}

// GetModule 获取模块
func (c *SysTeamAPI) GetModule() int32 {
	if c == nil {
		return 0
	}

	return c.Module
}

// GetDomain 获取领域
func (c *SysTeamAPI) GetDomain() int32 {
	if c == nil {
		return 0
	}

	return c.Domain
}

// String 字符串
func (c *SysTeamAPI) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis 存储实现
func (c *SysTeamAPI) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis 存储实现
func (c *SysTeamAPI) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName 表名
func (*SysTeamAPI) TableName() string {
	return tableNameSysTeamAPI
}
