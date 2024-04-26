package interflow

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aide-family/moon/pkg/helper/consts"
)

type (
	Callback func(topic consts.TopicType, value []byte) error

	HookMsg struct {
		// 投递任务
		Topic string `json:"topic"`
		// 投递数据
		Value []byte `json:"value"`
	}

	AgentInterflow interface {
		Interflow
		// Send 把数据投递给谁
		Send(ctx context.Context, msg *HookMsg) error

		// OnlineNotify 上线通知
		OnlineNotify() error
		// OfflineNotify 下线通知
		OfflineNotify() error
	}
)

const Timeout = 10 * time.Second

var receiveInterflowCh = make(chan *HookMsg, 100)

// GetReceiveInterflowCh 获取接收消息的通道
func GetReceiveInterflowCh() <-chan *HookMsg {
	return receiveInterflowCh
}

// GetSendInterflowCh 获取发送消息的通道
func GetSendInterflowCh() chan<- *HookMsg {
	return receiveInterflowCh
}

// Bytes send message to interflow
func (l *HookMsg) Bytes() []byte {
	tmp := l
	if tmp == nil {
		tmp = &HookMsg{}
	}
	bs, _ := json.Marshal(tmp)
	return bs
}
