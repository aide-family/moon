package msg

import (
	"encoding/json"

	"prometheus-manager/pkg/httpx"
)

var _ HookNotify = (*dingNotify)(nil)

type dingNotify struct{}

type (
	DingMsg struct {
		MsgType  string       `json:"msgtype"`
		Markdown DingMarkdown `json:"markdown"`
		At       DingAt       `json:"at"`
	}

	DingMarkdown struct {
		Title string `json:"title"`
		Text  string `json:"text"`
	}

	DingAt struct {
		AtMobiles []string `json:"atMobiles"`
		AtUserIds []string `json:"atUserIds"`
		IsAtAll   bool     `json:"isAtAll"`
	}
)

// Bytes DingMsg to bytes
func (v *DingMsg) Bytes() []byte {
	bs, _ := json.Marshal(v)
	return bs
}

func (l *dingNotify) Alarm(url string, msg *HookNotifyMsg) error {
	dingMsg := &DingMsg{
		MsgType: markdown,
		Markdown: DingMarkdown{
			Title: msg.Title,
			Text:  msg.Context,
		},
		At: DingAt{
			AtMobiles: []string{},
			AtUserIds: []string{},
			IsAtAll:   false,
		},
	}
	_, err := httpx.NewHttpX().POST(url, dingMsg.Bytes())
	return err
}

func NewDingNotify() HookNotify {
	return &dingNotify{}
}
