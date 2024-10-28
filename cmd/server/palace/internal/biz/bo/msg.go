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

type SendMsg struct {
	*hookapi.SendMsgRequest
}

func (s *SendMsg) Index() string {
	return s.RequestID
}
