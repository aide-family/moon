// Package other is a simple package that provides a other hook.
package other

import (
	"context"
	"fmt"
	"io"

	"github.com/aide-family/magicbox/config"
	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/httpx"
	"github.com/aide-family/magicbox/merr"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"github.com/aide-family/rabbit/pkg/message"
	"github.com/aide-family/rabbit/pkg/message/hook"
)

var _ message.Sender = (*otherHookSender)(nil)

func init() {
	message.RegisterDriver(enum.MessageType_WEBHOOK_OTHER, New)
}

func New(c *config.MessageConfig) (message.Sender, error) {
	options := &config.MessageWebhookConfig{}
	if err := anypb.UnmarshalTo(c.GetOptions(), options, proto.UnmarshalOptions{Merge: true}); err != nil {
		return nil, merr.ErrorInternalServer("unmarshal webhook config failed: %v", err)
	}
	return &otherHookSender{cli: httpx.NewClient(httpx.GetHTTPClient()), config: options}, nil
}

type otherHookSender struct {
	cli    *httpx.Client
	config *config.MessageWebhookConfig
}

// Send implements message.Sender.
func (o *otherHookSender) Send(ctx context.Context, message message.Message) error {
	if message.Type() != enum.MessageType_WEBHOOK_OTHER {
		return fmt.Errorf("message type %s not supported, only %s is supported", message.Type(), enum.MessageType_WEBHOOK_OTHER)
	}
	jsonBytes, err := message.Marshal()
	if err != nil {
		return err
	}

	resp, err := o.cli.Post(ctx, o.config.GetUrl(), jsonBytes)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return hook.RequestAssert(resp, unmarshalResponse)
}

func unmarshalResponse(body io.ReadCloser) error {
	return nil
}
