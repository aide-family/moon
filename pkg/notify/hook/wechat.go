package hook

import (
	"context"
	"fmt"
	"io"

	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/util/format"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/util/httpx"
)

var _ Notify = (*wechat)(nil)

// NewWechat 创建企业微信通知
func NewWechat(receiverHookWechatWork *conf.ReceiverHookWechatWork) Notify {
	return &wechat{
		ReceiverHookWechatWork: receiverHookWechatWork,
	}
}

type wechat struct {
	*conf.ReceiverHookWechatWork
}

type wechatHookResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *wechatHookResp) Error() string {
	return fmt.Sprintf("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
}

func (l *wechat) Send(ctx context.Context, msg notify.Msg) error {
	msgStr := l.GetTemplate()
	content := l.GetContent()
	if content != "" {
		msgStr = content
	}
	msgStr = format.Formatter(msgStr, msg)
	response, err := httpx.NewHTTPX().POSTWithContext(ctx, l.GetWebhook(), []byte(msgStr))
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
	var resp wechatHookResp
	if err = types.Unmarshal(resBytes, &resp); err != nil {
		return err
	}
	if resp.ErrCode != 0 {
		return &resp
	}
	return err
}
