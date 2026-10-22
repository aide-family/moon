package notification

import (
	"net"
	"net/mail"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/alerttmpl"
	"github.com/aide-family/Alioth/pkg/merr"
	pkgnotify "github.com/aide-family/Alioth/pkg/notify"
)

const (
	maxSubjectRunes    = 200
	maxBodyRunes       = 20000
	maxReceivers       = 50
	channelTypeSMTP    = "SMTP"
	channelTypeWebhook = "WEBHOOK"
	channelTypeSMS     = "SMS"
	appSMSAliCloud     = "SMS_ALICLOUD"
	templateTypeEmail  = "EMAIL"
	templateTypeHook   = "WEBHOOK"
	templateTypeSMS    = "SMS_ALICLOUD"
)

// channelApps 是渠道类型允许的应用。SMTP 统一用 OTHER 表示邮件；
// Webhook 区分厂商；SMS 只支持阿里云短信，发送实现尚未接入。
var channelApps = map[string]map[string]bool{
	channelTypeSMTP:    {"OTHER": true},
	channelTypeWebhook: {"OTHER": true, "DINGTALK": true, "WECHAT": true, "FEISHU": true},
	channelTypeSMS:     {appSMSAliCloud: true},
}

// webhookApps 是 Webhook 模板允许的应用。OTHER 是通用模板。
var webhookApps = map[string]bool{"OTHER": true, "DINGTALK": true, "WECHAT": true, "FEISHU": true}

var phonePattern = regexp.MustCompile(`^\+?[0-9]{6,20}$`)

// validateChannel 校验渠道类型、应用、地址和发件人。
func validateChannel(in bo.NotificationChannel) error {
	kind := strings.ToUpper(strings.TrimSpace(in.Type))
	app := strings.ToUpper(strings.TrimSpace(in.App))
	allowed, ok := channelApps[kind]
	if !ok {
		return merr.Invalid("invalid channel type")
	}
	if !allowed[app] {
		return merr.Invalid("channel type and app do not match")
	}
	endpoint := strings.TrimSpace(in.Endpoint)
	if endpoint == "" {
		return merr.Invalid("endpoint is required")
	}
	switch kind {
	case channelTypeSMTP:
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
	case channelTypeSMS:
		// 阿里云短信使用 HTTP API，地址校验和 Webhook 相同，但凭据和签名必填。
		if err := validateWebhookEndpoint(endpoint); err != nil {
			return err
		}
		if strings.TrimSpace(in.Username) == "" {
			return merr.Invalid("sms access key is required")
		}
		if strings.TrimSpace(in.FromAddress) == "" {
			return merr.Invalid("sms sign name is required")
		}
	default:
		if err := validateWebhookEndpoint(endpoint); err != nil {
			return err
		}
	}
	return nil
}

// validateTemplate 校验模板类型、应用、必填字段、长度和 Go template 语法。
func validateTemplate(in bo.NotificationTemplate) error {
	kind := strings.ToUpper(strings.TrimSpace(in.Type))
	app := strings.ToUpper(strings.TrimSpace(in.App))
	if app == "" {
		app = "OTHER"
	}
	switch kind {
	case templateTypeEmail:
		if app != "OTHER" {
			return merr.Invalid("email template app must be OTHER")
		}
		if strings.TrimSpace(in.Subject) == "" {
			return merr.Invalid("email subject is required")
		}
	case templateTypeSMS:
		if app != "OTHER" {
			return merr.Invalid("sms template app must be OTHER")
		}
	case templateTypeHook:
		if !webhookApps[app] {
			return merr.Invalid("invalid webhook template app")
		}
	default:
		return merr.Invalid("invalid template type")
	}
	subject := strings.TrimSpace(in.Subject)
	body := strings.TrimSpace(in.Body)
	if kind == templateTypeEmail && body == "" {
		return merr.Invalid("email body is required")
	}
	if (kind == templateTypeHook || kind == templateTypeSMS) && body == "" {
		return merr.Invalid("template body is required")
	}
	if utf8.RuneCountInString(subject) > maxSubjectRunes {
		return merr.Invalid("subject is too long")
	}
	if utf8.RuneCountInString(body) > maxBodyRunes {
		return merr.Invalid("body is too long")
	}
	// 用和投递相同的模板实现校验，Prometheus 的 $labels、$value 才算合法写法。
	if err := alerttmpl.Validate(in.Subject, "subject"); err != nil {
		return merr.Invalid("invalid subject template")
	}
	if err := alerttmpl.Validate(in.Body, "body"); err != nil {
		return merr.Invalid("invalid body template")
	}
	return nil
}

// validateGroupBinding 保证渠道和模板已启用，并且模板能覆盖该通知方式。
func validateGroupBinding(channel bo.NotificationChannel, tpl bo.NotificationTemplate) error {
	if channel.Status != "ACTIVE" {
		return merr.Invalid("channel is disabled")
	}
	if tpl.Status != "ACTIVE" {
		return merr.Invalid("template is disabled")
	}
	switch channel.Type {
	case channelTypeSMTP:
		if tpl.Type != templateTypeEmail {
			return merr.Invalid("smtp channel requires an email template")
		}
	case channelTypeSMS:
		if tpl.Type != templateTypeSMS {
			return merr.Invalid("sms channel requires an sms template")
		}
	case channelTypeWebhook:
		if tpl.Type != templateTypeHook {
			return merr.Invalid("webhook channel requires a webhook template")
		}
		// OTHER 是通用 Webhook 模板，可以覆盖任意 Webhook 应用；
		// 指定应用的模板只能覆盖同名应用，避免把企业微信模板发给钉钉。
		templateApp := tpl.App
		if templateApp == "" {
			templateApp = "OTHER"
		}
		if templateApp != "OTHER" && templateApp != channel.App {
			return merr.Invalid("template app does not cover the webhook app")
		}
	default:
		return merr.Invalid("invalid channel type")
	}
	return nil
}

// validateReceivers 按渠道类型校验接收方。邮件必须是非空邮箱列表，短信必须是手机号。
func validateReceivers(kind, raw string) error {
	switch strings.ToUpper(strings.TrimSpace(kind)) {
	case channelTypeSMTP:
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
	case channelTypeSMS:
		values := pkgnotify.SplitReceivers(raw)
		if len(values) == 0 {
			return merr.Invalid("sms receivers are required")
		}
		if len(values) > maxReceivers {
			return merr.Invalid("too many sms receivers")
		}
		for _, value := range values {
			if !phonePattern.MatchString(strings.TrimSpace(value)) {
				return merr.Invalid("invalid phone number")
			}
		}
	}
	return nil
}

func validatePhone(value string) error {
	if !phonePattern.MatchString(strings.TrimSpace(value)) {
		return merr.Invalid("invalid phone number")
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
