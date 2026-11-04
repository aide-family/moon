package repository

import (
	"context"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// EmailSettingRepo 维护登录验证码与邀请邮件使用的 SMTP 配置。
// 平台级单例；没有保存过时返回空配置，由调用方回落到服务器配置文件。
type EmailSettingRepo interface {
	// EmailSettings 读取 SMTP 配置；没有保存过时返回零值（Port 按 587），不返回 NotFound。
	EmailSettings(ctx context.Context) (*bo.EmailSettings, error)
	// SaveEmailSettings 覆盖保存 SMTP 配置。in.Password 为空且 clearPassword 为假时保留已存密码。
	SaveEmailSettings(ctx context.Context, in bo.EmailSettings, clearPassword bool) (*bo.EmailSettings, error)
}

// OAuthProviderSettingRepo 维护第三方登录来源的后台配置。
// 单个来源一行；没有保存过的来源由调用方回落到配置文件。
type OAuthProviderSettingRepo interface {
	// OAuthProviderSettings 列出已保存的来源配置，顺序不保证，由调用方按来源清单排序。
	OAuthProviderSettings(ctx context.Context) ([]bo.OAuthProviderSetting, error)
	// OAuthProviderSetting 读取单个来源；没有保存过时返回 NotFound。
	OAuthProviderSetting(ctx context.Context, provider string) (*bo.OAuthProviderSetting, error)
	// SaveOAuthProviderSetting 保存单个来源。in.ClientSecret 为空且 clearSecret 为假时保留已存密钥。
	SaveOAuthProviderSetting(ctx context.Context, in bo.OAuthProviderSetting, clearSecret bool) (*bo.OAuthProviderSetting, error)
}
