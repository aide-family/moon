// Package dingtalk implements the dingtalk hook driver.
package dingtalk

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/merr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/rabbit/pkg/message"
	"github.com/aide-family/rabbit/pkg/message/hook"
)

var _ message.Sender = (*dingtalkHookSender)(nil)

func init() {
	message.RegisterDriver(enum.MessageType_WEBHOOK_DINGTALK, New)
}

func New(c *config.MessageConfig) (message.Sender, error) {
	options := &config.MessageWebhookConfig{}
	if err := anypb.UnmarshalTo(c.GetOptions(), options, proto.UnmarshalOptions{Merge: true}); err != nil {
		return nil, merr.ErrorInternalServer("unmarshal webhook config failed: %v", err)
	}
	return &dingtalkHookSender{cli: httpx.NewClient(httpx.GetHTTPClient()), config: options}, nil
}

type dingtalkHookSender struct {
	cli    *httpx.Client
	config *config.MessageWebhookConfig
}

func (d *dingtalkHookSender) Send(ctx context.Context, message message.Message) error {
	if message.Type() != enum.MessageType_WEBHOOK_DINGTALK {
		return fmt.Errorf("message type %s not supported, only %s is supported", message.Type(), enum.MessageType_WEBHOOK_DINGTALK)
	}
	timestamp := time.Now().UnixMilli()
	opts := []httpx.Option{
		httpx.WithHeaders(map[string][]string{
			"Content-Type": {"application/json"},
		}),
		httpx.WithQuery(url.Values{
			// "access_token": {d.config.GetKey()},
			"timestamp": {strconv.FormatInt(timestamp, 10)},
			"sign":      {d.sign(timestamp)},
		}),
	}

	jsonBytes, err := message.Marshal()
	if err != nil {
		return err
	}

	resp, err := d.cli.Post(ctx, d.config.GetUrl(), jsonBytes, opts...)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return hook.RequestAssert(resp, unmarshalResponse)
}

func (d *dingtalkHookSender) sign(timestamp int64) string {
	message := fmt.Sprintf("%d\n%s", timestamp, d.config.GetSecret())
	key := []byte(d.config.GetSecret())
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	signature := h.Sum(nil)
	signatureBase64 := base64.StdEncoding.EncodeToString(signature)
	signatureURLEncoded := url.QueryEscape(signatureBase64)
	return signatureURLEncoded
}
