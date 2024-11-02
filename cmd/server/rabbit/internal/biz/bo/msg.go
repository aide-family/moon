package bo

import (
	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*SendMsgParams)(nil)

// SendMsgParams 发送消息请求参数
type SendMsgParams struct {
	Route     string
	Data      []byte
	RequestID string
}

func (s *SendMsgParams) String() string {
	return "rabbit:" + s.Route + ":" + s.RequestID
}

func (s *SendMsgParams) Key(app notify.Notify) string {
	return "rabbit:" + app.Type() + ":" + s.Route + ":" + types.MD5(s.RequestID+app.Hash())
}

func (s *SendMsgParams) Index() string {
	return s.RequestID
}

// Message 接收到的消息
func (s *SendMsgParams) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicAlertMsg)
}
