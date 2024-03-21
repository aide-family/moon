package msg

import (
	"context"
	"encoding/json"
	"fmt"
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

type WechatHookResp struct {
	ErrCode int    `json:"errcode"`
	ErrMsg  string `json:"errmsg"`
}

func (l *WechatHookResp) Error() string {
	return fmt.Sprintf("errcode: %d, errmsg: %s", l.ErrCode, l.ErrMsg)
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
	var resp WechatHookResp
	if err = json.Unmarshal(resBytes, &resp); err != nil {
		return err
	}
	if resp.ErrCode != 0 {
		return &resp
	}
	return err
}

func NewWechatNotify() HookNotify {
	return &wechatNotify{}
}
