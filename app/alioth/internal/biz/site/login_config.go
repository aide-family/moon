package site

import (
	"context"
	"net/mail"
	"strings"
	"unicode/utf8"

	"github.com/aide-family/moon/app/alioth/internal/mailer"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/oauthproviders"
)

const (
	maxSMTPHostRunes   = 255
	maxSMTPUserRunes   = 255
	maxOAuthClientID   = 256
	maxOAuthSecretRune = 512
)

// EmailSettings 读取后台配置的 SMTP，需要 system:login:read。
// 密码不回显：调用方只能拿到 PasswordSet，用来显示「已配置」。
func (settings *Settings) EmailSettings(ctx context.Context) (*bo.EmailSettings, error) {
	if _, err := settings.access.RequireSystem(ctx, perm.SystemLoginRead); err != nil {
		return nil, err
	}
	item, err := settings.emails.EmailSettings(ctx)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	item.Password = ""
	// 后台配置没启用（或没存主机）时，界面要显示服务器配置文件里当前生效的值，
	// 否则运营看到的是空白，以为邮件没配置。
	if !item.Enabled || item.Host == "" {
		if fallback := settings.configEmailSettings(); fallback != nil {
			passwordSet := item.PasswordSet || fallback.PasswordSet
			host, port := fallback.Host, fallback.Port
			from := fallback.From
			if item.Host != "" {
				host = item.Host
			}
			if item.Port > 0 {
				port = item.Port
			}
			if item.From != "" {
				from = item.From
			}
			return &bo.EmailSettings{
				Enabled:     item.Enabled,
				Host:        host,
				Port:        port,
				Username:    firstNonEmpty(item.Username, fallback.Username),
				Password:    "",
				PasswordSet: passwordSet,
				From:        from,
				Source:      bo.LoginConfigSourceConfig,
				UpdatedBy:   item.UpdatedBy,
				UpdatedAt:   item.UpdatedAt,
			}, nil
		}
		item.Source = bo.LoginConfigSourceNone
		return item, nil
	}
	item.Source = bo.LoginConfigSourceDatabase
	return item, nil
}

// UpdateEmailSettings 保存 SMTP 配置，需要 system:login:write。
// 保存成功后立即作用到发信器：关闭或清空主机时撤销覆盖，回落到服务器配置，
// 不需要重启服务。
func (settings *Settings) UpdateEmailSettings(ctx context.Context, in bo.EmailSettings, clearPassword bool) (*bo.EmailSettings, error) {
	account, err := settings.access.RequireSystem(ctx, perm.SystemLoginWrite)
	if err != nil {
		return nil, err
	}
	in.Host = strings.TrimSpace(in.Host)
	in.Username = strings.TrimSpace(in.Username)
	in.From = strings.TrimSpace(in.From)
	if err := validateEmailSettings(in); err != nil {
		return nil, err
	}
	in.UpdatedBy = account.UID
	saved, err := settings.emails.SaveEmailSettings(ctx, in, clearPassword)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	if settings.mail != nil {
		if saved.Enabled && saved.Host != "" {
			settings.mail.SetOverride(mailer.Config{
				Host: saved.Host, Port: saved.Port, Username: saved.Username,
				Password: saved.Password, From: saved.From,
			})
		} else {
			settings.mail.ClearOverride()
		}
	}
	saved.Password = ""
	return saved, nil
}

// OAuthProviders 按来源清单顺序返回后台配置，没有保存过的来源给空配置。
// 需要 system:login:read；密钥不回显，只给 SecretSet。
func (settings *Settings) OAuthProviders(ctx context.Context) ([]bo.OAuthProviderSetting, error) {
	if _, err := settings.access.RequireSystem(ctx, perm.SystemLoginRead); err != nil {
		return nil, err
	}
	stored, err := settings.oauths.OAuthProviderSettings(ctx)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	byProvider := make(map[string]bo.OAuthProviderSetting, len(stored))
	for _, item := range stored {
		byProvider[item.Provider] = item
	}
	items := make([]bo.OAuthProviderSetting, 0, len(oauthproviders.Catalog))
	for _, provider := range oauthproviders.Catalog {
		item := byProvider[provider.Key]
		item.Provider = provider.Key
		item.ClientSecret = ""
		if _, stored := byProvider[provider.Key]; stored {
			item.Source = bo.LoginConfigSourceDatabase
			items = append(items, item)
			continue
		}
		// 没有保存过就看配置文件：登录页此刻显示的就是这份，界面必须能看到。
		if fallback, ok := settings.configOAuthProvider(provider.Key); ok {
			items = append(items, fallback)
			continue
		}
		item.Source = bo.LoginConfigSourceNone
		items = append(items, item)
	}
	return items, nil
}

// UpdateOAuthProvider 保存单个来源，需要 system:login:write。
// 启用一个来源必须已经有 client id；密钥留空表示保留已存密钥。
func (settings *Settings) UpdateOAuthProvider(ctx context.Context, in bo.OAuthProviderSetting, clearSecret bool) (*bo.OAuthProviderSetting, error) {
	account, err := settings.access.RequireSystem(ctx, perm.SystemLoginWrite)
	if err != nil {
		return nil, err
	}
	in = normalizeOAuthProviderSetting(in)
	if err := validateOAuthProviderSetting(in); err != nil {
		return nil, err
	}
	in.UpdatedBy = account.UID
	saved, err := settings.oauths.SaveOAuthProviderSetting(ctx, in, clearSecret)
	if err != nil {
		return nil, settings.access.Wrap(err)
	}
	saved.ClientSecret = ""
	return saved, nil
}

// normalizeOAuthProviderSetting 统一来源标识大小写与两端空白。
func normalizeOAuthProviderSetting(in bo.OAuthProviderSetting) bo.OAuthProviderSetting {
	in.Provider = strings.ToUpper(strings.TrimSpace(in.Provider))
	in.ClientID = strings.TrimSpace(in.ClientID)
	return in
}

// validateOAuthProviderSetting 校验来源与凭据：来源必须在清单里，
// 启用一个来源必须已经填了 client id，否则登录页会出现点不动的按钮。
func validateOAuthProviderSetting(in bo.OAuthProviderSetting) error {
	if _, ok := oauthproviders.Lookup(in.Provider); !ok {
		return merr.Invalid("oauth provider is not supported")
	}
	if utf8.RuneCountInString(in.ClientID) > maxOAuthClientID || (in.Enabled && in.ClientID == "") {
		return merr.Invalid("invalid client id")
	}
	if utf8.RuneCountInString(in.ClientSecret) > maxOAuthSecretRune {
		return merr.Invalid("invalid client secret")
	}
	return nil
}

// configEmailSettings 把配置文件里的 SMTP 转成界面可展示的值（密码只标已配置）。
func (settings *Settings) configEmailSettings() *bo.EmailSettings {
	if settings.auth == nil || settings.auth.GetSmtp() == nil {
		return nil
	}
	smtp := settings.auth.GetSmtp()
	if strings.TrimSpace(smtp.GetHost()) == "" || smtp.GetPort() <= 0 || strings.TrimSpace(smtp.GetFrom()) == "" {
		return nil
	}
	return &bo.EmailSettings{
		Host:        smtp.GetHost(),
		Port:        int(smtp.GetPort()),
		Username:    smtp.GetUsername(),
		PasswordSet: smtp.GetPassword() != "",
		From:        smtp.GetFrom(),
	}
}

// configOAuthProvider 从配置文件里取某个来源的凭据（密钥只标已配置）。
func (settings *Settings) configOAuthProvider(provider string) (bo.OAuthProviderSetting, bool) {
	if settings.auth == nil {
		return bo.OAuthProviderSetting{}, false
	}
	for _, item := range settings.auth.GetOauth() {
		if !strings.EqualFold(item.GetProvider(), provider) {
			continue
		}
		if item.GetClientId() == "" || item.GetClientSecret() == "" {
			return bo.OAuthProviderSetting{}, false
		}
		return bo.OAuthProviderSetting{
			Provider:  provider,
			Enabled:   true,
			ClientID:  item.GetClientId(),
			SecretSet: true,
			Source:    bo.LoginConfigSourceConfig,
		}, true
	}
	return bo.OAuthProviderSetting{}, false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func validateEmailSettings(in bo.EmailSettings) error {
	if utf8.RuneCountInString(in.Host) > maxSMTPHostRunes {
		return merr.Invalid("invalid smtp host")
	}
	if utf8.RuneCountInString(in.Username) > maxSMTPUserRunes {
		return merr.Invalid("invalid smtp host")
	}
	if in.Port < 0 || in.Port > 65535 {
		return merr.Invalid("invalid smtp port")
	}
	if in.From != "" {
		if _, err := mail.ParseAddress(in.From); err != nil {
			return merr.Invalid("invalid smtp from")
		}
	}
	// 启用时必须能发信：主机、端口与发件人缺一个都会让验证码发不出去。
	if in.Enabled {
		if in.Host == "" || in.Port <= 0 {
			return merr.Invalid("invalid smtp port")
		}
		if in.From == "" {
			return merr.Invalid("invalid smtp from")
		}
	}
	return nil
}
