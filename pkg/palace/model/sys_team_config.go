package model

import (
	"github.com/aide-family/moon/pkg/util/cipher"
	"github.com/aide-family/moon/pkg/util/email"
	"github.com/aide-family/moon/pkg/util/types"
)

const tableNameSysTeamConfig = "sys_team_config"

// SysTeamConfig mapped from table <sys_team_config>
type SysTeamConfig struct {
	AllFieldModel
	TeamID      uint32               `gorm:"column:team_id;type:int unsigned;not null;index:sys_teams__sys_team_email,unique,priority:1;comment:团队id" json:"team_id"`
	EmailConfig *email.DefaultConfig `gorm:"column:email_config;type:text;not null;comment:邮箱配置" json:"email_config"`
	// 对称加密配置
	SymmetricEncryptionConfig *cipher.SymmetricEncryptionConfig `gorm:"column:symmetric_encryption;type:text;not null;comment:对称加密配置" json:"symmetric_encryption_config"`
	// 非对称加密配置
	AsymmetricEncryptionConfig *cipher.AsymmetricEncryptionConfig `gorm:"column:asymmetric_encryption;type:text;not null;comment:非对称加密配置" json:"asymmetric_encryption_config"`
}

// String json string of SysTeamConfig
func (c *SysTeamConfig) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

// UnmarshalBinary redis存储实现
func (c *SysTeamConfig) UnmarshalBinary(data []byte) error {
	return types.Unmarshal(data, c)
}

// MarshalBinary redis存储实现
func (c *SysTeamConfig) MarshalBinary() (data []byte, err error) {
	return types.Marshal(c)
}

// TableName SysTeamConfig's table name
func (*SysTeamConfig) TableName() string {
	return tableNameSysTeamConfig
}
