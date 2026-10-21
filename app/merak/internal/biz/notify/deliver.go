package notify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/aide-family/Alioth/app/merak/pkg/bo"
	"github.com/aide-family/Alioth/pkg/alerttmpl"
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

// render 用任务参数渲染模板。告警摘要、描述和通知模板共用 pkg/alerttmpl 的
// 数据和函数，所以 $labels、$value、.Labels、.Value、{{ . }} 两边含义一致。
// 解析或执行失败时返回错误，避免把未替换的模板当成成功内容发出去。
func render(raw, paramsJSON string) (string, error) {
	if raw == "" {
		return "", nil
	}
	params := map[string]any{}
	if paramsJSON != "" {
		if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
			return "", fmt.Errorf("template params invalid: %w", err)
		}
	}
	data := alerttmpl.FromParams(params)
	externalURL, _ := url.Parse(stringValue(data["ExternalURL"]))
	result, err := alerttmpl.Render(context.Background(), raw, data, "notify", time.Now(), unavailableQuery, externalURL)
	if err != nil {
		return "", fmt.Errorf("template render failed: %w", err)
	}
	return result, nil
}

// unavailableQuery 说明 query 模板函数依赖数据源，只在评测阶段可用。
func unavailableQuery(context.Context, string, time.Time) ([]alerttmpl.Sample, error) {
	return nil, errors.New("query is not available in notification templates")
}

func stringValue(value any) string {
	text, _ := value.(string)
	return text
}
