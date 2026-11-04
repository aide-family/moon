package site

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// 启用自定义 SMTP 时必须给全主机、端口和发件人，否则验证码发不出去。
func TestValidateEmailSettingsRequiresFieldsWhenEnabled(t *testing.T) {
	cases := []bo.EmailSettings{
		{Enabled: true, Port: 587, From: "a@b.com"},
		{Enabled: true, Host: "smtp.example.com", Port: 0, From: "a@b.com"},
		{Enabled: true, Host: "smtp.example.com", Port: 587},
	}
	for _, in := range cases {
		if err := validateEmailSettings(in); err == nil {
			t.Fatalf("incomplete smtp config should be rejected: %+v", in)
		}
	}
	if err := validateEmailSettings(bo.EmailSettings{
		Enabled: true, Host: "smtp.example.com", Port: 587, From: "Alioth <a@b.com>",
	}); err != nil {
		t.Fatalf("complete smtp config should be accepted: %v", err)
	}
}

// 关闭自定义 SMTP 时允许字段为空：此时回落到服务器配置。
func TestValidateEmailSettingsAllowsEmptyWhenDisabled(t *testing.T) {
	if err := validateEmailSettings(bo.EmailSettings{Enabled: false}); err != nil {
		t.Fatalf("disabled smtp config should be accepted: %v", err)
	}
}

func TestValidateEmailSettingsRejectsBadInputs(t *testing.T) {
	if err := validateEmailSettings(bo.EmailSettings{Port: 70000}); err == nil {
		t.Fatal("port above 65535 should be rejected")
	}
	if err := validateEmailSettings(bo.EmailSettings{From: "not-a-mailbox"}); err == nil {
		t.Fatal("invalid from address should be rejected")
	}
}

// 来源必须在清单里；启用却填不出 client id 的配置要挡下来，
// 否则登录页会出现一个点不动的第三方按钮。
func TestValidateOAuthProviderSetting(t *testing.T) {
	if err := validateOAuthProviderSetting(bo.OAuthProviderSetting{Provider: "GITHUB", Enabled: true}); err == nil {
		t.Fatal("enabled provider without client id should be rejected")
	}
	if err := validateOAuthProviderSetting(bo.OAuthProviderSetting{Provider: "WEIBO"}); err == nil {
		t.Fatal("unknown provider should be rejected")
	}
	if err := validateOAuthProviderSetting(bo.OAuthProviderSetting{
		Provider: "GITHUB", Enabled: true, ClientID: "id", ClientSecret: "secret",
	}); err != nil {
		t.Fatalf("complete provider config should be accepted: %v", err)
	}
}

func TestNormalizeOAuthProviderSetting(t *testing.T) {
	got := normalizeOAuthProviderSetting(bo.OAuthProviderSetting{Provider: " github ", ClientID: " id "})
	if got.Provider != "GITHUB" || got.ClientID != "id" {
		t.Fatalf("provider config not normalized: %+v", got)
	}
}
