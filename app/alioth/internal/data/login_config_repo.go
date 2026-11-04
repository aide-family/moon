package data

import (
	"context"
	"strings"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/authemailsetting"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/authoauthprovider"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/pkg/merr"
)

// emailSettingUID 是 SMTP 配置单例行的固定标识。
const emailSettingUID = "email"

const defaultSMTPPort = 587

func (r *emailSettingRepo) EmailSettings(ctx context.Context) (*bo.EmailSettings, error) {
	row, err := r.client.AuthEmailSetting.Query().
		Where(authemailsetting.UIDEQ(emailSettingUID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			// 没有保存过就是「不使用后台配置」，发信回落到服务器配置。
			return &bo.EmailSettings{Port: defaultSMTPPort}, nil
		}
		return nil, r.fail(err)
	}
	return r.emailSettingsOf(row)
}

func (r *emailSettingRepo) SaveEmailSettings(ctx context.Context, in bo.EmailSettings, clearPassword bool) (*bo.EmailSettings, error) {
	cipher, err := r.emailPasswordCipher(ctx, in.Password, clearPassword)
	if err != nil {
		return nil, err
	}
	_, err = r.client.AuthEmailSetting.Query().
		Where(authemailsetting.UIDEQ(emailSettingUID)).
		Only(ctx)
	if ent.IsNotFound(err) {
		_, err = r.client.AuthEmailSetting.Create().
			SetUID(emailSettingUID).
			SetEnabled(in.Enabled).
			SetHost(in.Host).
			SetPort(in.Port).
			SetUsername(in.Username).
			SetPasswordCipher(cipher).
			SetFrom(in.From).
			SetUpdatedBy(in.UpdatedBy).
			Save(ctx)
	} else if err == nil {
		_, err = r.client.AuthEmailSetting.Update().
			Where(authemailsetting.UIDEQ(emailSettingUID)).
			SetEnabled(in.Enabled).
			SetHost(in.Host).
			SetPort(in.Port).
			SetUsername(in.Username).
			SetPasswordCipher(cipher).
			SetFrom(in.From).
			SetUpdatedBy(in.UpdatedBy).
			Save(ctx)
	}
	if err != nil {
		return nil, r.fail(err)
	}
	return r.EmailSettings(ctx)
}

// emailPasswordCipher 决定这次保存写入什么密码：显式清空写空串，
// 传了新的就加密，什么都没传就保留已存密文，避免「只想改端口却把密码抹掉」。
func (r *emailSettingRepo) emailPasswordCipher(ctx context.Context, plain string, clear bool) (string, error) {
	if clear {
		return "", nil
	}
	if plain != "" {
		cipher, err := r.box.encrypt(plain)
		if err != nil {
			return "", r.fail(err)
		}
		return cipher, nil
	}
	row, err := r.client.AuthEmailSetting.Query().
		Where(authemailsetting.UIDEQ(emailSettingUID)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", r.fail(err)
	}
	return row.PasswordCipher, nil
}

func (r *emailSettingRepo) emailSettingsOf(row *ent.AuthEmailSetting) (*bo.EmailSettings, error) {
	password, err := r.box.decrypt(row.PasswordCipher)
	if err != nil {
		return nil, r.fail(err)
	}
	return &bo.EmailSettings{
		Enabled:     row.Enabled,
		Host:        row.Host,
		Port:        row.Port,
		Username:    row.Username,
		Password:    password,
		PasswordSet: row.PasswordCipher != "",
		From:        row.From,
		UpdatedBy:   row.UpdatedBy,
		UpdatedAt:   row.UpdatedAt,
	}, nil
}

func (r *oauthProviderSettingRepo) OAuthProviderSettings(ctx context.Context) ([]bo.OAuthProviderSetting, error) {
	rows, err := r.client.AuthOAuthProvider.Query().All(ctx)
	if err != nil {
		return nil, r.fail(err)
	}
	items := make([]bo.OAuthProviderSetting, 0, len(rows))
	for _, row := range rows {
		item, err := r.oauthProviderSettingOf(row)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}
	return items, nil
}

func (r *oauthProviderSettingRepo) OAuthProviderSetting(ctx context.Context, provider string) (*bo.OAuthProviderSetting, error) {
	row, err := r.client.AuthOAuthProvider.Query().
		Where(authoauthprovider.ProviderEQ(strings.ToUpper(provider))).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, merr.NotFound("oauth provider is not configured")
		}
		return nil, r.fail(err)
	}
	return r.oauthProviderSettingOf(row)
}

func (r *oauthProviderSettingRepo) SaveOAuthProviderSetting(ctx context.Context, in bo.OAuthProviderSetting, clearSecret bool) (*bo.OAuthProviderSetting, error) {
	provider := strings.ToUpper(strings.TrimSpace(in.Provider))
	cipher, err := r.oauthSecretCipher(ctx, provider, in.ClientSecret, clearSecret)
	if err != nil {
		return nil, err
	}
	existing, err := r.client.AuthOAuthProvider.Query().
		Where(authoauthprovider.ProviderEQ(provider)).
		IDs(ctx)
	if ent.IsNotFound(err) || (err == nil && len(existing) == 0) {
		_, err = r.client.AuthOAuthProvider.Create().
			SetUID(id.New()).
			SetProvider(provider).
			SetEnabled(in.Enabled).
			SetClientID(in.ClientID).
			SetClientSecretCipher(cipher).
			SetUpdatedBy(in.UpdatedBy).
			Save(ctx)
	} else if err == nil {
		_, err = r.client.AuthOAuthProvider.Update().
			Where(authoauthprovider.ProviderEQ(provider)).
			SetEnabled(in.Enabled).
			SetClientID(in.ClientID).
			SetClientSecretCipher(cipher).
			SetUpdatedBy(in.UpdatedBy).
			Save(ctx)
	}
	if err != nil {
		return nil, r.fail(err)
	}
	return r.OAuthProviderSetting(ctx, provider)
}

// oauthSecretCipher 与 SMTP 密码同一套语义：显式清空写空串，传了新的就加密，
// 没传就保留已存密文。
func (r *oauthProviderSettingRepo) oauthSecretCipher(ctx context.Context, provider, plain string, clear bool) (string, error) {
	if clear {
		return "", nil
	}
	if plain != "" {
		cipher, err := r.box.encrypt(plain)
		if err != nil {
			return "", r.fail(err)
		}
		return cipher, nil
	}
	row, err := r.client.AuthOAuthProvider.Query().
		Where(authoauthprovider.ProviderEQ(provider)).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return "", nil
		}
		return "", r.fail(err)
	}
	return row.ClientSecretCipher, nil
}

func (r *oauthProviderSettingRepo) oauthProviderSettingOf(row *ent.AuthOAuthProvider) (*bo.OAuthProviderSetting, error) {
	secret, err := r.box.decrypt(row.ClientSecretCipher)
	if err != nil {
		return nil, r.fail(err)
	}
	return &bo.OAuthProviderSetting{
		Provider:     row.Provider,
		Enabled:      row.Enabled,
		ClientID:     row.ClientID,
		ClientSecret: secret,
		SecretSet:    row.ClientSecretCipher != "",
		UpdatedBy:    row.UpdatedBy,
		UpdatedAt:    row.UpdatedAt,
	}, nil
}
