package notification

import (
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestValidateChannel(t *testing.T) {
	valid := bo.NotificationChannel{
		Type: "SMTP", App: "OTHER", Endpoint: "smtp.example.com:587",
		Username: "sender@example.com", FromAddress: "sender@example.com",
	}
	if err := validateChannel(valid); err != nil {
		t.Fatalf("valid smtp rejected: %v", err)
	}
	invalid := []bo.NotificationChannel{
		{Type: "SMTP", App: "DINGTALK", Endpoint: "smtp.example.com:587", Username: "u"},
		{Type: "SMTP", App: "OTHER", Endpoint: "smtp.example.com:abc", Username: "u"},
		{Type: "SMTP", App: "OTHER", Endpoint: "", Username: "u"},
		{Type: "SMTP", App: "OTHER", Endpoint: "smtp.example.com:587"},
		{Type: "WEBHOOK", App: "SMS_ALICLOUD", Endpoint: "https://example.com/hook"},
		{Type: "WEBHOOK", App: "OTHER", Endpoint: "ftp://example.com/hook"},
		{Type: "WEBHOOK", App: "OTHER", Endpoint: "https://user:pass@example.com/hook"},
	}
	for _, item := range invalid {
		if err := validateChannel(item); err == nil {
			t.Fatalf("invalid channel accepted: %+v", item)
		}
	}
	webhook := bo.NotificationChannel{Type: "WEBHOOK", App: "DINGTALK", Endpoint: "https://example.com/hook", Secret: "s"}
	if err := validateChannel(webhook); err != nil {
		t.Fatalf("valid webhook rejected: %v", err)
	}
	sms := bo.NotificationChannel{
		Type: "SMS", App: "SMS_ALICLOUD", Endpoint: "https://dysmsapi.aliyuncs.com",
		Username: "ak", Password: "sk", FromAddress: "Alioth",
	}
	if err := validateChannel(sms); err != nil {
		t.Fatalf("valid sms rejected: %v", err)
	}
	if err := validateChannel(bo.NotificationChannel{Type: "SMS", App: "SMS_ALICLOUD", Endpoint: "https://dysmsapi.aliyuncs.com"}); err == nil {
		t.Fatal("sms without credentials accepted")
	}
}

func TestValidateTemplate(t *testing.T) {
	if err := validateTemplate(bo.NotificationTemplate{Type: "EMAIL", Subject: "告警", Body: "{{.summary}}"}); err != nil {
		t.Fatalf("valid email rejected: %v", err)
	}
	if err := validateTemplate(bo.NotificationTemplate{Type: "WEBHOOK", Body: "{{.summary}}"}); err != nil {
		t.Fatalf("valid webhook rejected: %v", err)
	}
	if err := validateTemplate(bo.NotificationTemplate{Type: "WEBHOOK", App: "WECHAT", Body: "{{.summary}}"}); err != nil {
		t.Fatalf("valid wechat webhook rejected: %v", err)
	}
	if err := validateTemplate(bo.NotificationTemplate{Type: "SMS_ALICLOUD", Body: "【{{.phase}}】{{.summary}}"}); err != nil {
		t.Fatalf("valid sms template rejected: %v", err)
	}
	if err := validateTemplate(bo.NotificationTemplate{Type: "WEBHOOK", App: "WECHAT", Body: "{{.summary}}", Subject: "{{.summary}}"}); err != nil {
		t.Fatalf("valid wechat webhook with subject rejected: %v", err)
	}
	// 通知模板和告警摘要、描述共用模板实现，Prometheus 写法必须合法。
	if err := validateTemplate(bo.NotificationTemplate{
		Type: "WEBHOOK", Body: "【{{ .phase }}】{{ $labels.job }} {{ humanize $value }} {{ . }}",
	}); err != nil {
		t.Fatalf("prometheus template rejected: %v", err)
	}
	invalid := []bo.NotificationTemplate{
		{Type: "EMAIL", Body: "body"},
		{Type: "EMAIL", Subject: "subject"},
		{Type: "EMAIL", Subject: "subject", Body: "{{.summary"},
		{Type: "WEBHOOK", Body: "{{ $labels.job "},
		{Type: "WEBHOOK", Body: ""},
		{Type: "WEBHOOK", App: "SMS_ALICLOUD", Body: "b"},
		{Type: "SMS_ALICLOUD", Body: ""},
		{Type: "EMAIL", App: "WECHAT", Subject: "s", Body: "b"},
	}
	for _, item := range invalid {
		if err := validateTemplate(item); err == nil {
			t.Fatalf("invalid template accepted: %+v", item)
		}
	}
}

func TestValidateGroupBinding(t *testing.T) {
	smtp := bo.NotificationChannel{Type: "SMTP", Status: "ACTIVE"}
	webhook := bo.NotificationChannel{Type: "WEBHOOK", Status: "ACTIVE"}
	email := bo.NotificationTemplate{Type: "EMAIL", Status: "ACTIVE"}
	web := bo.NotificationTemplate{Type: "WEBHOOK", App: "OTHER", Status: "ACTIVE"}
	smsChannel := bo.NotificationChannel{Type: "SMS", App: "SMS_ALICLOUD", Status: "ACTIVE"}
	smsTemplate := bo.NotificationTemplate{Type: "SMS_ALICLOUD", App: "OTHER", Status: "ACTIVE"}
	if err := validateGroupBinding(smtp, email); err != nil {
		t.Fatalf("smtp+email rejected: %v", err)
	}
	if err := validateGroupBinding(webhook, web); err != nil {
		t.Fatalf("webhook+webhook rejected: %v", err)
	}
	if err := validateGroupBinding(smsChannel, smsTemplate); err != nil {
		t.Fatalf("sms+sms rejected: %v", err)
	}
	if err := validateGroupBinding(smsChannel, web); err == nil {
		t.Fatal("sms+webhook template accepted")
	}
	wechat := bo.NotificationChannel{Type: "WEBHOOK", App: "WECHAT", Status: "ACTIVE"}
	wechatTemplate := bo.NotificationTemplate{Type: "WEBHOOK", App: "WECHAT", Status: "ACTIVE"}
	if err := validateGroupBinding(wechat, wechatTemplate); err != nil {
		t.Fatalf("wechat+wechat template rejected: %v", err)
	}
	dingTalkTemplate := bo.NotificationTemplate{Type: "WEBHOOK", App: "DINGTALK", Status: "ACTIVE"}
	if err := validateGroupBinding(wechat, dingTalkTemplate); err == nil {
		t.Fatal("wechat+dingtalk template accepted")
	}
	if err := validateGroupBinding(smtp, web); err == nil {
		t.Fatal("smtp+webhook template accepted")
	}
	if err := validateGroupBinding(webhook, email); err == nil {
		t.Fatal("webhook+email template accepted")
	}
	disabled := bo.NotificationChannel{Type: "SMTP", Status: "DISABLED"}
	if err := validateGroupBinding(disabled, email); err == nil {
		t.Fatal("disabled channel accepted")
	}
}

func TestValidateReceivers(t *testing.T) {
	if err := validateReceivers("SMTP", "a@example.com, b@example.com"); err != nil {
		t.Fatalf("valid receivers rejected: %v", err)
	}
	if err := validateReceivers("SMTP", ""); err == nil {
		t.Fatal("empty smtp receivers accepted")
	}
	if err := validateReceivers("SMTP", "not-an-email"); err == nil {
		t.Fatal("invalid email accepted")
	}
	if err := validateReceivers("WEBHOOK", ""); err != nil {
		t.Fatalf("empty webhook receivers rejected: %v", err)
	}
	if err := validateReceivers("SMS", "13800138000,+8613900139000"); err != nil {
		t.Fatalf("valid phones rejected: %v", err)
	}
	if err := validateReceivers("SMS", "not-a-phone"); err == nil {
		t.Fatal("invalid phone accepted")
	}
	if err := validateReceivers("SMS", ""); err == nil {
		t.Fatal("empty sms receivers accepted")
	}
}
