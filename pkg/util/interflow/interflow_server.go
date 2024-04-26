package interflow

import (
	"context"

	"github.com/aide-family/moon/pkg/helper/consts"
)

type (
	Interflow interface {
		// Receive 接收投递过来的数据
		Receive() error
		// SetHandles 设置回调函数
		SetHandles(handles map[consts.TopicType]Callback) error
		// Close 关闭
		Close() error
	}

	ServerInterflow interface {
		Interflow
		// SendAgent 把数据投递给谁
		SendAgent(ctx context.Context, to string, msg *HookMsg) error

		// ServerOnlineNotify 上线通知
		ServerOnlineNotify(agentUrls []string) error
		// ServerOfflineNotify 下线通知
		ServerOfflineNotify(agentUrls []string) error
	}
)
