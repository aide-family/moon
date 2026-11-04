package bo

import "time"

// EmailSettings 是后台配置的 SMTP 发信参数，平台级单例。
// Enabled 为假或没有配置主机时，发信回落到服务器配置文件里的 SMTP。
type EmailSettings struct {
	Enabled  bool
	Host     string
	Port     int
	Username string
	// Password 只在 biz 内部使用：接口只回显 PasswordSet，不回显明文。
	Password    string
	PasswordSet bool
	From        string
	// Source 表示这份值从哪来：DATABASE（后台保存过且启用）、
	// CONFIG（回落到服务器配置）、NONE（哪里都没配）。
	Source    string
	UpdatedBy string
	UpdatedAt time.Time
}

// OAuthProviderSetting 是单个第三方登录来源的后台配置。
// ClientSecret 只在 biz 内部使用：接口只回显 SecretSet。
type OAuthProviderSetting struct {
	Provider     string
	Enabled      bool
	ClientID     string
	ClientSecret string
	SecretSet    bool
	// Source 表示这份值从哪来：DATABASE、CONFIG 或 NONE。
	Source    string
	UpdatedBy string
	UpdatedAt time.Time
}

// 登录配置的来源标识。
const (
	LoginConfigSourceDatabase = "DATABASE"
	LoginConfigSourceConfig   = "CONFIG"
	LoginConfigSourceNone     = "NONE"
)
