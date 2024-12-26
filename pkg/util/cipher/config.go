package cipher

import (
	"database/sql/driver"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewSymmetricEncryptionConfig 创建对称加密配置
func NewSymmetricEncryptionConfig(key, iv string) *SymmetricEncryptionConfig {
	return &SymmetricEncryptionConfig{Key: key, Iv: iv}
}

// NewAsymmetricEncryptionConfig 创建非对称加密配置
func NewAsymmetricEncryptionConfig(publicKey, privateKey string) *AsymmetricEncryptionConfig {
	return &AsymmetricEncryptionConfig{PublicKey: publicKey, PrivateKey: privateKey}
}

type (
	// SymmetricEncryptionConfig 对称加密配置
	SymmetricEncryptionConfig struct {
		// 密钥
		Key string `json:"key"`
		// 初始化向量
		Iv string `json:"iv"`
	}

	// AsymmetricEncryptionConfig 非对称加密配置
	AsymmetricEncryptionConfig struct {
		// 公钥
		PublicKey string `json:"public_key"`
		// 私钥
		PrivateKey string `json:"private_key"`
	}
)

// Scan 实现gorm的Scan方法
func (c *SymmetricEncryptionConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return types.Unmarshal(v, &c)
	case string:
		return types.Unmarshal([]byte(v), &c)
	default:
		return merr.ErrorNotificationSystemError("invalid type")
	}
}

// Value 实现gorm的Value方法
func (c *SymmetricEncryptionConfig) Value() (driver.Value, error) {
	if types.IsNil(c) {
		return []byte("{}"), nil
	}
	return types.Marshal(c)
}

// Scan 实现gorm的Scan方法
func (c *AsymmetricEncryptionConfig) Scan(value interface{}) error {
	switch v := value.(type) {
	case []byte:
		return types.Unmarshal(v, &c)
	case string:
		return types.Unmarshal([]byte(v), &c)
	default:
		return merr.ErrorNotificationSystemError("invalid type")
	}
}

// Value 实现gorm的Value方法
func (c *AsymmetricEncryptionConfig) Value() (driver.Value, error) {
	if types.IsNil(c) {
		return []byte("{}"), nil
	}
	return types.Marshal(c)
}

// ToConf 转换为conf.SymmetricEncryptionConfig
func (c *SymmetricEncryptionConfig) ToConf() *conf.SymmetricEncryptionConfig {
	if types.IsNil(c) {
		return nil
	}
	return &conf.SymmetricEncryptionConfig{
		Key: c.Key,
		Iv:  c.Iv,
	}
}

// ToConf 转换为conf.AsymmetricEncryptionConfig
func (c *AsymmetricEncryptionConfig) ToConf() *conf.AsymmetricEncryptionConfig {
	if types.IsNil(c) {
		return nil
	}
	return &conf.AsymmetricEncryptionConfig{
		PublicKey:  c.PublicKey,
		PrivateKey: c.PrivateKey,
	}
}
