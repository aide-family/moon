// Package feishu is a simple package that provides a feishu hook.
package feishu

import (
	"context"
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

	feishuMessage := &Message{}
	var ok bool
	if feishuMessage, ok = message.(*Message); !ok {
		jsonBytes, err := message.Marshal()
		if err != nil {
			return err
		}
		feishuMessage = &Message{}
		if err := json.Unmarshal(jsonBytes, feishuMessage); err != nil {
			return err
		}
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	feishuMessage.Timestamp = timestamp
	if err := feishuMessage.Signature(f.config.GetSecret()); err != nil {
		return err
	}

	u, err := url.Parse(f.config.GetUrl())
	if err != nil {
		return err
	}

	jsonBytes, err := feishuMessage.Marshal()
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
