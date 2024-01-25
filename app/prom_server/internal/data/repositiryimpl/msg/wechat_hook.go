package msg

import (
	"encoding/json"

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

func (l *wechatNotify) Alarm(url string, msg *HookNotifyMsg) error {
	wechatMsg := &WechatMsg{
		MsgType:  markdown,
		Markdown: WechatMarkdown{Content: msg.Content},
	}
	_, err := httpx.NewHttpX().POST(url, wechatMsg.Bytes())
	return err
}

func NewWechatNotify() HookNotify {
	return &wechatNotify{}
}
