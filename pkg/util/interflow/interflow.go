package interflow

import (
	"context"

	"prometheus-manager/pkg/helper/consts"
)

type (
	Callback func(topic consts.TopicType, key, value []byte) error

	Interflow interface {
		// Send 把数据投递给谁
		Send(ctx context.Context, topic string, key []byte, value []byte) error
		// Receive 接收投递过来的数据
		Receive() error
		// SetHandles 设置回调函数
		SetHandles(handles map[consts.TopicType]Callback) error
		// Close 关闭
		Close() error
	}
)
