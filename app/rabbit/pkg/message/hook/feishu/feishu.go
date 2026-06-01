// Package feishu is a simple package that provides a feishu hook.
package feishu

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/merr"
	klog "github.com/go-kratos/kratos/v2/log"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/rabbit/pkg/message"
	"github.com/aide-family/rabbit/pkg/message/hook"
)

var _ message.Sender = (*feishuHookSender)(nil)

func init() {
	message.RegisterDriver(enum.MessageType_WEBHOOK_FEISHU, New)
}

func New(c *config.MessageConfig) (message.Sender, error) {
	options := &config.MessageWebhookConfig{}
	if err := anypb.UnmarshalTo(c.GetOptions(), options, proto.UnmarshalOptions{Merge: true}); err != nil {
		return nil, merr.ErrorInternalServer("unmarshal webhook config failed: %v", err)
	}
	return &feishuHookSender{cli: httpx.NewClient(httpx.GetHTTPClient()), config: options}, nil
}

type feishuHookSender struct {
	cli    *httpx.Client
	config *config.MessageWebhookConfig
}

// Send implements message.Sender.
func (f *feishuHookSender) Send(ctx context.Context, message message.Message) error {
	if message.Type() != enum.MessageType_WEBHOOK_FEISHU {
		return fmt.Errorf("message type %s not supported, only %s is supported", message.Type(), enum.MessageType_WEBHOOK_FEISHU)
	}
	opts := []httpx.Option{
		httpx.WithHeaders(map[string][]string{
			"Content-Type": {"application/json"},
		}),
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	var jsonBytes []byte
	var err error

	if feishuMessage, ok := message.(*Message); ok {
		feishuMessage.Timestamp = timestamp
		if err := feishuMessage.Signature(f.config.GetSecret()); err != nil {
			return err
		}
		jsonBytes, err = feishuMessage.Marshal()
	} else {
		rawBody, err := message.Marshal()
		if err != nil {
			return err
		}
		jsonBytes, err = signFeishuWebhookPayload(rawBody, timestamp, f.config.GetSecret())
	}
	if err != nil {
		return err
	}

	u, err := url.Parse(f.config.GetUrl())
	if err != nil {
		return err
	}

	klog.Debugf("feishu message: %s", string(jsonBytes))
	resp, err := f.cli.Post(ctx, u.String(), jsonBytes, opts...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return hook.RequestAssert(resp, unmarshalResponse)
}

func signFeishuWebhookPayload(rawBody []byte, timestamp, secret string) ([]byte, error) {
	sign, err := feishuWebhookSign(timestamp, secret)
	if err != nil {
		return nil, err
	}
	if len(rawBody) == 0 {
		return nil, fmt.Errorf("feishu message body is empty")
	}

	var payload map[string]json.RawMessage
	if err := json.Unmarshal(rawBody, &payload); err != nil {
		return nil, err
	}
	payload["timestamp"] = json.RawMessage(strconv.Quote(timestamp))
	payload["sign"] = json.RawMessage(strconv.Quote(sign))
	return json.Marshal(payload)
}

func feishuWebhookSign(timestamp, secret string) (string, error) {
	signString := timestamp + "\n" + secret
	h := hmac.New(sha256.New, []byte(signString))
	if _, err := h.Write(nil); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
}
