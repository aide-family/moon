package interflow

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/consts"
)

type (
	Callback func(topic consts.TopicType, key, value []byte) error

	HookMsg struct {
		Topic string `json:"topic"`
		Value []byte `json:"value"`
		Key   []byte `json:"key"`
	}

	Interflow interface {
		// Send 把数据投递给谁
		Send(ctx context.Context, to string, msg *HookMsg) error
		// Receive 接收投递过来的数据
		Receive() error
		// SetHandles 设置回调函数
		SetHandles(handles map[consts.TopicType]Callback) error
		// Close 关闭
		Close() error
	}
)
