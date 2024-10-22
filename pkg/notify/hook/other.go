package hook

import (
	"context"
	"io"

	"github.com/aide-family/moon/pkg/util/format"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/util/httpx"
)

var _ Notify = (*other)(nil)

// NewOther 创建企业微信通知
func NewOther(config Config) Notify {
	return &other{
		c: config,
	}
}

type other struct {
	c Config
}

func (l *other) Type() string {
	return l.c.GetType()
}

func (l *other) Send(ctx context.Context, msg notify.Msg) error {
	msgStr := l.c.GetTemplate()
	content := l.c.GetContent()
	if content != "" {
		msgStr = content
	}
	msgStr = format.Formatter(msgStr, msg)
	response, err := httpx.NewHTTPX().POSTWithContext(ctx, l.c.GetWebhook(), []byte(msgStr))
	if err != nil {
		return err
	}
	body := response.Body
	defer body.Close()
	resBytes, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	return err
}
