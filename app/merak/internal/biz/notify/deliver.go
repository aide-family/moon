package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/aide-family/Alioth/app/merak/pkg/bo"
)

// sender 是某一类渠道的投递实现。每种 Webhook 应用有自己的签名和响应格式。
type sender interface {
	send(ctx context.Context, task bo.TaskSnapshot, subject, body string) error
}

// DeliveryResult 包含这次投递尝试实际使用的渲染内容，供日志审计。
type DeliveryResult struct {
	OK      bool
	Detail  string
	Subject string
	Content string
}

// Deliver 渲染模板并投递。短信类型直接失败，不报 SENT。
func Deliver(ctx context.Context, task bo.TaskSnapshot) DeliveryResult {
	if task.ChannelApp == "SMS_ALICLOUD" || task.TemplateType == "SMS_ALICLOUD" {
		return DeliveryResult{Detail: "sms delivery is not implemented", Subject: task.Subject, Content: task.Body}
	}
	subject := task.Subject
	body := task.Body
	if subject == "" {
		subject = task.TemplateSubject
	}
	if body == "" {
		body = task.TemplateBody
	}
	subject, err := render(subject, task.Params)
	if err != nil {
		return DeliveryResult{Detail: err.Error(), Subject: subject, Content: body}
	}
	body, err = render(body, task.Params)
	if err != nil {
		return DeliveryResult{Detail: err.Error(), Subject: subject, Content: body}
	}
	result := DeliveryResult{Subject: subject, Content: body}
	impl, err := senderFor(task.ChannelType, task.ChannelApp)
	if err != nil {
		result.Detail = err.Error()
		return result
	}
	if err := impl.send(ctx, task, subject, body); err != nil {
		result.Detail = err.Error()
		return result
	}
	result.OK = true
	result.Detail = "sent"
	return result
}

func senderFor(kind, app string) (sender, error) {
	switch kind {
	case "SMTP":
		return smtpSender{}, nil
	case "WEBHOOK":
		switch app {
		case "OTHER":
			return otherSender{}, nil
		case "DINGTALK":
			return dingTalkSender{}, nil
		case "WECHAT":
			return wechatSender{}, nil
		case "FEISHU":
			return feishuSender{}, nil
		default:
			return nil, fmt.Errorf("unsupported webhook app %s", app)
		}
	default:
		return nil, fmt.Errorf("unsupported channel type %s", kind)
	}
}

// render 用任务参数渲染模板。解析或执行失败时返回错误，避免把未替换的模板当成成功内容发出去。
func render(raw, paramsJSON string) (string, error) {
	if raw == "" {
		return "", nil
	}
	data := map[string]any{}
	if paramsJSON != "" {
		_ = json.Unmarshal([]byte(paramsJSON), &data)
	}
	tpl, err := template.New("notify").Option("missingkey=zero").Parse(raw)
	if err != nil {
		return "", fmt.Errorf("template parse failed: %w", err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("template render failed: %w", err)
	}
	return buf.String(), nil
}
