package msg

import (
	"encoding/json"

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

func (l *wechatNotify) Alarm(url string, msg *HookNotifyMsg) error {
	wechatMsg := &WechatMsg{
		MsgType:  markdown,
		Markdown: WechatMarkdown{Content: msg.Content},
	}
	response, err := httpx.NewHttpX().POST(url, wechatMsg.Bytes())
	var resBytes []byte
	body := response.Body
	for {
		tmp := make([]byte, 0, 4096)
		read, err := body.Read(tmp)
		if err != nil {
			return err
		}
		resBytes = append(resBytes, tmp...)
		if read == 0 {
			break
		}
	}
	log.Infow("notify", string(resBytes))
	return err
}

func NewWechatNotify() HookNotify {
	return &wechatNotify{}
}
