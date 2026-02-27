// Package wechat is a simple package that provides a wechat hook.
package wechat

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/merr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/rabbit/pkg/message"
	"github.com/aide-family/rabbit/pkg/message/hook"
)

var _ message.Sender = (*wechatHookSender)(nil)

func init() {
	message.RegisterDriver(enum.MessageType_WEBHOOK_WECHAT, New)
}

func New(c *config.MessageConfig) (message.Sender, error) {
	options := &config.MessageWebhookConfig{}
	if err := anypb.UnmarshalTo(c.GetOptions(), options, proto.UnmarshalOptions{Merge: true}); err != nil {
		return nil, merr.ErrorInternalServer("unmarshal webhook config failed: %v", err)
	}
	return &wechatHookSender{config: options}, nil
}

type wechatHookSender struct {
	config *config.MessageWebhookConfig
}

// Send implements message.Sender.
func (w *wechatHookSender) Send(ctx context.Context, msg message.Message) error {
	if msg.Type() != enum.MessageType_WEBHOOK_WECHAT {
		return fmt.Errorf("message type %s not supported, only %s is supported", msg.Type(), enum.MessageType_WEBHOOK_WECHAT)
	}
	opts := []httpx.Option{
		httpx.WithHeaders(map[string][]string{
			"Content-Type": {"application/json"},
		}),
	}
	if secret := w.config.GetSecret(); secret != "" {
		opts = append(opts, httpx.WithQuery(url.Values{
			"key": {secret},
		}))
	}
	newMessage := &Message{}
	var ok bool
	if newMessage, ok = msg.(*Message); !ok {
		jsonBytes, err := msg.Marshal()
		if err != nil {
			return err
		}
		newMessage = &Message{}
		if err := json.Unmarshal(jsonBytes, newMessage); err != nil {
			return err
		}
	}
	jsonBytes, err := newMessage.Marshal()
	if err != nil {
		return err
	}
	client := httpx.NewClient(httpx.GetHTTPClient())
	resp, err := client.Post(ctx, w.config.GetUrl(), jsonBytes, opts...)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return hook.RequestAssert(resp, unmarshalResponse)
}
