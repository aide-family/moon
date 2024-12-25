package email

import (
	"database/sql/driver"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
)

var _ Config = (*DefaultConfig)(nil)

// DefaultConfig 默认邮件配置
type DefaultConfig struct {
	User string
	Pass string
	Host string
	Port uint32
}

// GetHost implements Config.
func (d *DefaultConfig) GetHost() string {
	if d == nil {
		return ""
	}
	return d.Host
}

// GetPass implements Config.
func (d *DefaultConfig) GetPass() string {
	if d == nil {
		return ""
	}
	return d.Pass
}

// GetPort implements Config.
func (d *DefaultConfig) GetPort() uint32 {
	if d == nil {
		return 0
	}
	return d.Port
}

// 实现email.Config接口
func (d *DefaultConfig) GetUser() string {
	if d == nil {
		return ""
	}
	return d.User
}

// Scan 实现gorm的Scan方法
func (d *DefaultConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return types.Unmarshal(v, d)
	case string:
		return types.Unmarshal([]byte(v), d)
	default:
		return merr.ErrorNotificationSystemError("invalid type")
	}
}

// Value 实现gorm的Value方法
func (d *DefaultConfig) Value() (driver.Value, error) {
	return types.Marshal(d)
}

// ToConf 转换为conf.EmailConfig
func (d *DefaultConfig) ToConf() *conf.EmailConfig {
	if types.IsNil(d) {
		return nil
	}
	return &conf.EmailConfig{
		User: d.User,
		Pass: d.Pass,
		Host: d.Host,
		Port: d.Port,
	}
}

// NewDefaultConfig 创建默认邮件配置
func NewDefaultConfig(config *conf.EmailConfig) *DefaultConfig {
	return &DefaultConfig{
		User: config.GetUser(),
		Pass: config.GetPass(),
		Host: config.GetHost(),
		Port: config.GetPort(),
	}
}
