package bo

import (
	hookapi "github.com/aide-family/moon/api/rabbit/hook"
	"github.com/aide-family/moon/pkg/watch"
)

type (
	// Message 消息明细
	Message struct {
		Data map[string]any
	}
)

var _ watch.Indexer = (*SendMsg)(nil)

// SendMsg 发送消息
type SendMsg struct {
	*hookapi.SendMsgRequest
}

// Index 生成发送消息索引
func (s *SendMsg) Index() string {
	return s.RequestID
}
