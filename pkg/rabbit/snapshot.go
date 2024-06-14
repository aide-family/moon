package rabbit

import (
	"context"
	"fmt"
)

// MessageSnapshot 消息快照
// 用于在处理消息时生成一份不可改变的数据，后续的进程都将基于这个快照进行处理
// 快照生成失败，则该消息会重回队列，等待下一次处理
type MessageSnapshot struct {
	Context    context.Context
	CancelFunc context.CancelFunc
	Message    *Message
	// Processors 包含消息所有处理需要的处理器，进行处理时，直接使用即可
	Processors map[int64]*Processor
}

type Processor struct {
	Index      string
	Secret     []byte
	Suppressor Suppressor
	Templater  Templater
	Sender     Sender
}

func NewMessageSnapshot(ctx context.Context, message *Message) *MessageSnapshot {
	snapCtx, cancelFunc := context.WithCancel(ctx)

	return &MessageSnapshot{
		Context:    snapCtx,
		CancelFunc: cancelFunc,
		Message:    message,
	}
}

// CompleteMessageSnapshot 补全快照信息。
func (m *MessageSnapshot) CompleteMessageSnapshot(ctx context.Context, configGetter ConfigGetter) error {
	var runners = make(map[int64]*Processor, len(m.Message.Templates))
	for _, id := range m.Message.Templates {
		var (
			templater  Templater
			suppressor Suppressor
			sender     Sender
			secret     []byte
			err        error
		)
		templater, err = configGetter.GetTemplater(ctx, id)
		if err != nil {
			return err
		}
		suppressor, err = configGetter.GetSuppressorByTemplate(ctx, id)
		if err != nil {
			return err
		}
		sender, err = configGetter.GetSenderByTemplate(ctx, id)
		if err != nil {
			return err
		}
		secret, err = configGetter.GetSecret(ctx, id)
		if err != nil {
			return err
		}
		runners[id] = &Processor{
			Index:      IndexFunc(id, m.Message.Group),
			Suppressor: suppressor,
			Templater:  templater,
			Sender:     sender,
			Secret:     secret,
		}
	}
	m.Processors = runners
	return nil
}

func IndexFunc(tid int64, group string) string {
	return fmt.Sprintf("%d-%s", tid, group)
}
