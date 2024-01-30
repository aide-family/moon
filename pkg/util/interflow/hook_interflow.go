package interflow

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/after"
	"prometheus-manager/pkg/helper/consts"
	"prometheus-manager/pkg/httpx"
)

var _ Interflow = (*hookInterflow)(nil)

var receiveInterflowCh = make(chan *HookMsg, 100)

type (
	hookInterflow struct {
		log     *log.Helper
		handles map[consts.TopicType]Callback
		lock    sync.RWMutex
		closeCh chan struct{}
	}
)

// GetSendInterflowCh 获取发送消息的通道
func GetSendInterflowCh() chan<- *HookMsg {
	return receiveInterflowCh
}

// GetReceiveInterflowCh 获取接收消息的通道
func GetReceiveInterflowCh() <-chan *HookMsg {
	return receiveInterflowCh
}

// Bytes send message to interflow
func (l *HookMsg) Bytes() []byte {
	if l == nil {
		return []byte("{}")
	}
	bs, _ := json.Marshal(l)
	return bs
}

func (l *hookInterflow) Close() error {
	close(l.closeCh)
	return nil
}

func (l *hookInterflow) Send(ctx context.Context, to string, msg *HookMsg) error {
	_, err := httpx.NewHttpX().POST(to, msg.Bytes())
	if err != nil {
		l.log.Errorw("err", err, "key", msg.Key, "topic", msg.Topic, "value", string(msg.Value))
		return err
	}
	return nil
}

func (l *hookInterflow) Receive() error {
	receiveCh := GetReceiveInterflowCh()
	go func() {
		defer after.Recover(l.log)
		for {
			select {
			case msg := <-receiveCh:
				if handle, ok := l.handles[consts.TopicType(msg.Topic)]; ok {
					if err := handle(consts.TopicType(msg.Topic), msg.Key, msg.Value); err != nil {
						l.log.Errorw("err", err, "topic", msg.Topic, "value", string(msg.Value), "key", string(msg.Key))
					}
				}
			case <-l.closeCh:
				l.log.Info("hookInterflow closed")
				return
			}
		}
	}()
	return nil
}

func (l *hookInterflow) SetHandles(handles map[consts.TopicType]Callback) error {
	l.lock.Lock()
	defer l.lock.Unlock()
	l.handles = handles
	return nil
}

func NewHookInterflow(log *log.Helper) Interflow {
	return &hookInterflow{
		log:     log,
		closeCh: make(chan struct{}),
	}
}
