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

// String 转换为字符串
func (s *SendMsgParams) String() string {
	return types.TextJoin("rabbit:", s.Route, ":", s.RequestID)
}

// Key 获取消息的键
func (s *SendMsgParams) Key(app notify.Notify) string {
	return types.TextJoin("rabbit:", app.Type(), ":", s.Route, ":", types.MD5(types.TextJoin(s.RequestID, app.Hash())))
}

// Index 获取消息的索引
func (s *SendMsgParams) Index() string {
	return s.RequestID
}

// Message 接收到的消息
func (s *SendMsgParams) Message() *watch.Message {
	return watch.NewMessage(s, vobj.TopicAlertMsg)
}
