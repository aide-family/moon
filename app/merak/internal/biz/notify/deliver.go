package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/smtp"
	"strings"
	"text/template"

	"github.com/aide-family/Alioth/app/merak/pkg/bo"
)

// Deliver 渲染模板并投递。短信类型直接失败，不报 SENT。
func Deliver(ctx context.Context, task bo.TaskSnapshot) (bool, string) {
	if task.ChannelApp == "SMS_ALICLOUD" || task.TemplateType == "SMS_ALICLOUD" {
		return false, "sms delivery is not implemented"
	}
	subject := task.Subject
	body := task.Body
	if subject == "" {
		subject = task.TemplateSubject
	}
	if body == "" {
		body = task.TemplateBody
	}
	subject = render(subject, task.Params)
	body = render(body, task.Params)
	switch task.ChannelType {
	case "SMTP":
		if err := sendSMTP(ctx, task, subject, body); err != nil {
			return false, err.Error()
		}
		return true, "sent"
	case "WEBHOOK":
		if err := sendWebhook(ctx, task, subject, body); err != nil {
			return false, err.Error()
		}
		return true, "sent"
	default:
		return false, "unsupported channel type"
	}
}

func render(raw, paramsJSON string) string {
	if raw == "" {
		return ""
	}
	data := map[string]any{}
	if paramsJSON != "" {
		_ = json.Unmarshal([]byte(paramsJSON), &data)
	}
	tpl, err := template.New("notify").Option("missingkey=zero").Parse(raw)
	if err != nil {
		return raw
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return raw
	}
	return buf.String()
}

func sendSMTP(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	host, port, err := net.SplitHostPort(task.ChannelEndpoint)
	if err != nil {
		host = task.ChannelEndpoint
		port = "25"
	}
	addr := net.JoinHostPort(host, port)
	from := task.ChannelFrom
	if from == "" {
		from = task.ChannelUsername
	}
	receivers := splitReceivers(task.Receivers)
	if len(receivers) == 0 {
		return fmt.Errorf("no receivers")
	}
	msg := []byte("From: " + from + "\r\nTo: " + strings.Join(receivers, ",") + "\r\nSubject: " + subject + "\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n" + body)
	var auth smtp.Auth
	if task.ChannelUsername != "" {
		auth = smtp.PlainAuth("", task.ChannelUsername, task.ChannelPassword, host)
	}
	dialer := &net.Dialer{}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	client, err := smtp.NewClient(conn, host)
	if err != nil {
		_ = conn.Close()
		return err
	}
	defer client.Close()
	if auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	for _, to := range receivers {
		if err := client.Rcpt(to); err != nil {
			return err
		}
	}
	writer, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := writer.Write(msg); err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}
	return client.Quit()
}

func sendWebhook(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	payload, contentType := webhookPayload(task.ChannelApp, subject, body, task.ChannelSecret)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, task.ChannelEndpoint, bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", contentType)
	if task.ChannelSecret != "" && task.ChannelApp == "OTHER" {
		req.Header.Set("X-Signature", task.ChannelSecret)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 300 {
		return fmt.Errorf("webhook status %d", resp.StatusCode)
	}
	return nil
}

func webhookPayload(app, subject, body, secret string) ([]byte, string) {
	text := body
	if subject != "" {
		text = subject + "\n" + body
	}
	switch app {
	case "DINGTALK":
		raw, _ := json.Marshal(map[string]any{"msgtype": "text", "text": map[string]string{"content": text}})
		return raw, "application/json"
	case "WECHAT":
		raw, _ := json.Marshal(map[string]any{"msgtype": "text", "text": map[string]string{"content": text}})
		return raw, "application/json"
	case "FEISHU":
		raw, _ := json.Marshal(map[string]any{"msg_type": "text", "content": map[string]string{"text": text}})
		return raw, "application/json"
	default:
		raw, _ := json.Marshal(map[string]string{"subject": subject, "body": body, "secret": secret})
		return raw, "application/json"
	}
}

func splitReceivers(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ';' || r == '\n'
	})
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}
