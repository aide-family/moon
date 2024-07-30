package rabbit

import (
	"context"

	"github.com/aide-family/moon/api"
)

// MessageSnapshot 消息快照
// 用于在处理消息时生成一份不可改变的数据，后续的进程都将基于这个快照进行处理
// 快照生成失败，则该消息会重回队列，等待下一次处理
type MessageSnapshot struct {
	Context    context.Context
	CancelFunc context.CancelFunc
	Message    *api.Message
	// Processors 包含消息所有处理需要的处理器，进行处理时，直接使用即可
	Processors map[string]*Processor
}

// Processor 处理器
type Processor struct {
	Filter     Filter
	Aggregator Aggregator
	Templater  Templater
	Sender     Sender
}

// NewMessageSnapshot 创建消息快照。
func NewMessageSnapshot(ctx context.Context, message *api.Message) *MessageSnapshot {
	snapCtx, cancelFunc := context.WithCancel(ctx)

	return &MessageSnapshot{
		Context:    snapCtx,
		CancelFunc: cancelFunc,
		Message:    message,
	}
}

// CompleteMessageSnapshot 补全快照信息。
func (m *MessageSnapshot) CompleteMessageSnapshot(ctx context.Context, rp RuleGroupProvider, pp ProcessorProvider) error {
	var runners = make(map[string]*Processor, len(m.Message.UseGroups))
	for _, name := range m.Message.UseGroups {
		var (
			filter     Filter
			aggregator Aggregator
			templater  Templater
			sender     Sender
			err        error
		)
		rg, err := rp.RuleGroup(ctx, name)
		if err != nil {
			return err
		}
		filter, err = pp.Filter(ctx, rg.FilterRuleName)
		if err != nil {
			return err
		}
		aggregator, err = pp.Aggregator(ctx, rg.AggregationRuleName)
		if err != nil {
			return err
		}
		templater, err = pp.Templater(ctx, rg.TemplateRuleName)
		if err != nil {
			return err
		}
		sender, err = pp.Sender(ctx, rg.SendRuleName)
		if err != nil {
			return err
		}

		runners[name] = &Processor{
			Filter:     filter,
			Aggregator: aggregator,
			Templater:  templater,
			Sender:     sender,
		}
	}
	m.Processors = runners
	return nil
}
