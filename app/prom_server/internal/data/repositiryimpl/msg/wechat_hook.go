package msg

import (
	"context"
	"encoding/json"
	"io"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/httpx"
)

var _ HookNotify = (*wechatNotify)(nil)

type wechatNotify struct{}

type WechatMsg struct {
	MsgType  string         `json:"msgtype"`
	Markdown WechatMarkdown `json:"markdown"`
}

type WechatMarkdown struct {
	Content string `json:"content"`
}

// Bytes WechatMsg to bytes
func (v *WechatMsg) Bytes() []byte {
	bs, _ := json.Marshal(v)
	return bs
}

func (l *wechatNotify) Alarm(ctx context.Context, url string, msg *HookNotifyMsg) error {
	response, err := httpx.NewHttpX().POSTWithContext(ctx, url, []byte(msg.Content))
	body := response.Body
	resBytes, err := io.ReadAll(body)
	defer body.Close()
	if err != nil {
		return err
	}
	log.Debugw("notify", string(resBytes))
	return err
}

func NewWechatNotify() HookNotify {
	return &wechatNotify{}
}
