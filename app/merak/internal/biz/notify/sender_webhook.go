package notify

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/aide-family/Alioth/app/merak/pkg/bo"
)

const maxWebhookResponse = 64 << 10

var webhookClient = &http.Client{Timeout: 30 * time.Second}

type otherSender struct{}

func (otherSender) send(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	payload, err := json.Marshal(map[string]string{"subject": subject, "body": body})
	if err != nil {
		return err
	}
	headers := map[string]string{}
	if task.ChannelSecret != "" {
		timestamp := strconv.FormatInt(time.Now().Unix(), 10)
		headers["X-Alioth-Timestamp"] = timestamp
		headers["X-Alioth-Signature"] = "sha256=" + hex.EncodeToString(hmacSHA256([]byte(task.ChannelSecret), []byte(timestamp+"."+string(payload))))
	}
	_, err = postRaw(ctx, task.ChannelEndpoint, payload, headers)
	return err
}

type dingTalkSender struct{}

func (dingTalkSender) send(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	endpoint := task.ChannelEndpoint
	if task.ChannelSecret != "" {
		timestamp := time.Now().UnixMilli()
		values := url.Values{}
		values.Set("timestamp", strconv.FormatInt(timestamp, 10))
		values.Set("sign", dingTalkSign(task.ChannelSecret, timestamp))
		updated, err := appendQuery(endpoint, values)
		if err != nil {
			return err
		}
		endpoint = updated
	}
	payload := map[string]any{"msgtype": "text", "text": map[string]string{"content": webhookText(subject, body)}}
	data, err := postJSON(ctx, endpoint, payload)
	if err != nil {
		return err
	}
	var reply struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &reply); err != nil {
		return fmt.Errorf("invalid dingtalk response")
	}
	if reply.ErrCode != 0 {
		return fmt.Errorf("dingtalk errcode %d: %s", reply.ErrCode, reply.ErrMsg)
	}
	return nil
}

func dingTalkSign(secret string, timestamp int64) string {
	stringToSign := fmt.Sprintf("%d\n%s", timestamp, secret)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

type wechatSender struct{}

func (wechatSender) send(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	payload := map[string]any{"msgtype": "text", "text": map[string]string{"content": webhookText(subject, body)}}
	data, err := postJSON(ctx, task.ChannelEndpoint, payload)
	if err != nil {
		return err
	}
	var reply struct {
		ErrCode int    `json:"errcode"`
		ErrMsg  string `json:"errmsg"`
	}
	if err := json.Unmarshal(data, &reply); err != nil {
		return fmt.Errorf("invalid wechat response")
	}
	if reply.ErrCode != 0 {
		return fmt.Errorf("wechat errcode %d: %s", reply.ErrCode, reply.ErrMsg)
	}
	return nil
}

type feishuSender struct{}

func (feishuSender) send(ctx context.Context, task bo.TaskSnapshot, subject, body string) error {
	payload := map[string]any{
		"msg_type": "text",
		"content":  map[string]string{"text": webhookText(subject, body)},
	}
	if task.ChannelSecret != "" {
		timestamp := time.Now().Unix()
		payload["timestamp"] = strconv.FormatInt(timestamp, 10)
		payload["sign"] = feishuSign(task.ChannelSecret, timestamp)
	}
	data, err := postJSON(ctx, task.ChannelEndpoint, payload)
	if err != nil {
		return err
	}
	var reply struct {
		Code          int    `json:"code"`
		Msg           string `json:"msg"`
		StatusCode    int    `json:"StatusCode"`
		StatusMessage string `json:"StatusMessage"`
	}
	if err := json.Unmarshal(data, &reply); err != nil {
		return fmt.Errorf("invalid feishu response")
	}
	if reply.Code != 0 {
		return fmt.Errorf("feishu code %d: %s", reply.Code, reply.Msg)
	}
	if reply.StatusCode != 0 {
		return fmt.Errorf("feishu status %d: %s", reply.StatusCode, reply.StatusMessage)
	}
	return nil
}

func feishuSign(secret string, timestamp int64) string {
	stringToSign := strconv.FormatInt(timestamp, 10) + "\n" + secret
	mac := hmac.New(sha256.New, []byte(stringToSign))
	return base64.StdEncoding.EncodeToString(mac.Sum(nil))
}

func webhookText(subject, body string) string {
	if subject == "" {
		return body
	}
	return subject + "\n" + body
}

func postJSON(ctx context.Context, endpoint string, payload any) ([]byte, error) {
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return postRaw(ctx, endpoint, raw, nil)
}

func postRaw(ctx context.Context, endpoint string, body []byte, headers map[string]string) ([]byte, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return nil, fmt.Errorf("invalid webhook endpoint")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, parsed.String(), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	for key, value := range headers {
		req.Header.Set(key, value)
	}
	resp, err := webhookClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxWebhookResponse+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxWebhookResponse {
		return nil, fmt.Errorf("webhook response is too large")
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("webhook status %d: %s", resp.StatusCode, truncate(string(data), 200))
	}
	return data, nil
}

func appendQuery(endpoint string, values url.Values) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", fmt.Errorf("invalid webhook endpoint")
	}
	query := parsed.Query()
	for key, items := range values {
		for _, item := range items {
			query.Set(key, item)
		}
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func hmacSHA256(key, message []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(message)
	return mac.Sum(nil)
}

func truncate(value string, limit int) string {
	value = strings.TrimSpace(value)
	if len(value) <= limit {
		return value
	}
	return value[:limit]
}
