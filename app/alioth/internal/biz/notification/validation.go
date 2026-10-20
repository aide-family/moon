package notification

import (
	"net"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"text/template"
	"unicode/utf8"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/merr"
	pkgnotify "github.com/aide-family/Alioth/pkg/notify"
)

const (
	maxSubjectRunes    = 200
	maxBodyRunes       = 20000
	maxReceivers       = 50
	channelTypeSMTP    = "SMTP"
	channelTypeWebhook = "WEBHOOK"
	appSMSAliCloud     = "SMS_ALICLOUD"
)

// channelApps 是渠道类型允许的应用。SMTP 统一用 OTHER 表示邮件；Webhook 才区分厂商。
var channelApps = map[string]map[string]bool{
	channelTypeSMTP:    {"OTHER": true},
	channelTypeWebhook: {"OTHER": true, "DINGTALK": true, "WECHAT": true, "FEISHU": true},
}

// validateChannel 校验渠道类型、应用、地址和发件人。
func validateChannel(in bo.NotificationChannel) error {
	kind := strings.ToUpper(strings.TrimSpace(in.Type))
	app := strings.ToUpper(strings.TrimSpace(in.App))
	allowed, ok := channelApps[kind]
	if !ok {
		return merr.Invalid("invalid channel type")
	}
	if app == appSMSAliCloud {
		return merr.Invalid("sms delivery is not implemented")
	}
	if !allowed[app] {
		return merr.Invalid("channel type and app do not match")
	}
	endpoint := strings.TrimSpace(in.Endpoint)
	if endpoint == "" {
		return merr.Invalid("endpoint is required")
	}
	if kind == channelTypeSMTP {
		if err := validateSMTPEndpoint(endpoint); err != nil {
			return err
		}
		if strings.TrimSpace(in.FromAddress) == "" && strings.TrimSpace(in.Username) == "" {
			return merr.Invalid("from address or username is required")
		}
		if strings.TrimSpace(in.FromAddress) != "" {
			if _, err := mail.ParseAddress(strings.TrimSpace(in.FromAddress)); err != nil {
				return merr.Invalid("invalid from address")
			}
		}
		return nil
	}
	if err := validateWebhookEndpoint(endpoint); err != nil {
		return err
	}
	return nil
}

// validateTemplate 校验模板类型、必填字段、长度和 Go template 语法。
func validateTemplate(in bo.NotificationTemplate) error {
	kind := strings.ToUpper(strings.TrimSpace(in.Type))
	switch kind {
	case "EMAIL":
		if strings.TrimSpace(in.Subject) == "" {
			return merr.Invalid("email subject is required")
		}
	case "WEBHOOK":
	case appSMSAliCloud:
		return merr.Invalid("sms delivery is not implemented")
	default:
		return merr.Invalid("invalid template type")
	}
	subject := strings.TrimSpace(in.Subject)
	body := strings.TrimSpace(in.Body)
	if kind == "EMAIL" && body == "" {
		return merr.Invalid("email body is required")
	}
	if kind == "WEBHOOK" && body == "" {
		return merr.Invalid("webhook body is required")
	}
	if utf8.RuneCountInString(subject) > maxSubjectRunes {
		return merr.Invalid("subject is too long")
	}
	if utf8.RuneCountInString(body) > maxBodyRunes {
		return merr.Invalid("body is too long")
	}
	if _, err := template.New("subject").Parse(in.Subject); err != nil {
		return merr.Invalid("invalid subject template")
	}
	if _, err := template.New("body").Parse(in.Body); err != nil {
		return merr.Invalid("invalid body template")
	}
	return nil
}

// validateGroupBinding 保证渠道、模板已启用且类型匹配。
func validateGroupBinding(channel bo.NotificationChannel, tpl bo.NotificationTemplate) error {
	if channel.Status != "ACTIVE" {
		return merr.Invalid("channel is disabled")
	}
	if tpl.Status != "ACTIVE" {
		return merr.Invalid("template is disabled")
	}
	if channel.Type == channelTypeSMTP && tpl.Type != "EMAIL" {
		return merr.Invalid("smtp channel requires an email template")
	}
	if channel.Type == channelTypeWebhook && tpl.Type != "WEBHOOK" {
		return merr.Invalid("webhook channel requires a webhook template")
	}
	return nil
}

// validateReceivers 按渠道类型校验接收方。SMTP 必须是非空邮箱列表，Webhook 不要求。
func validateReceivers(kind, raw string) error {
	if strings.ToUpper(strings.TrimSpace(kind)) != channelTypeSMTP {
		return nil
	}
	values := pkgnotify.SplitReceivers(raw)
	if len(values) == 0 {
		return merr.Invalid("receivers are required")
	}
	if len(values) > maxReceivers {
		return merr.Invalid("too many receivers")
	}
	for _, value := range values {
		if _, err := mail.ParseAddress(value); err != nil {
			return merr.Invalid("invalid receiver address")
		}
	}
	return nil
}

func validateSMTPEndpoint(value string) error {
	if strings.Contains(value, "://") || strings.ContainsAny(value, "/ \t\r\n") {
		return merr.Invalid("invalid smtp endpoint")
	}
	host, port, err := net.SplitHostPort(value)
	if err != nil {
		// 没有端口时使用默认 25，host 仍然要做基本校验。
		host = value
		port = "25"
	}
	if strings.TrimSpace(host) == "" {
		return merr.Invalid("invalid smtp endpoint")
	}
	number, err := strconv.Atoi(port)
	if err != nil || number < 1 || number > 65535 {
		return merr.Invalid("invalid smtp port")
	}
	return nil
}

func validateWebhookEndpoint(value string) error {
	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return merr.Invalid("invalid webhook endpoint")
	}
	if parsed.User != nil {
		return merr.Invalid("webhook endpoint must not contain credentials")
	}
	return nil
}
