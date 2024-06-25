package rabbit

import (
	"context"
	"io"

	"github.com/aide-family/moon/api"
)

// Plugin 仅仅自定义一个Name的接口
// 整体来看，每个模块都是一个插件，它们都是可插拔的
type Plugin interface {
	Name() string
}

// Receiver 用来接收来自外部消息
type Receiver interface {
	Plugin
	Receive() (<-chan *api.Message, error)
}

// Filter 按照规则对消息进行过滤
type Filter interface {
	Plugin
	Inject(rule Rule) (Filter, error)
	Allow(message *api.Message) bool
}

// Aggregator 按照规则对消息进行聚合，聚合完成则返回消息
type Aggregator interface {
	Plugin
	Inject(rule Rule) (Aggregator, error)
	Group(in *api.Message) (out *api.Message, err error)
}

// Templater 负责对消息进行模版的解析。
type Templater interface {
	Plugin
	Inject(rule Rule) (Templater, error)
	Parse(in any, out io.Writer) error
}

// Sender 负责发送消息，它不需要关心什么时候发送，也不需要关心消息的内容是什么。
// 它只负责在接收到消息时，将它按照已经确定的方式发送出去。
type Sender interface {
	Plugin
	Inject(rule Rule) (Sender, error)
	Send(ctx context.Context, content []byte) error
}

// ConfigProvider Sender 发送消息时，有的需要 Config 才能够将送达。
// 然而 Sender 并不需要关心 Config 如何正确生成，它只需调用 Provider 接口，获取到 Config 使用即可。
// 之所以这样设计，是因为:
// + 不同的 Sender 需要的 Config 的数据结构不同
// + 相同的 Sender 不同的接收者需要的 Config 不同
type ConfigProvider interface {
	// Provider 负责在被调用时提供正确的密钥，否则返回错误
	Provider(in []byte, out any) error
}

type RuleGroupProvider interface {
	RuleGroup(ctx context.Context, name string) (*RuleGroup, error)
}

type ProcessorProvider interface {
	Filter(ctx context.Context, ruleName string) (Filter, error)
	Aggregator(ctx context.Context, ruleName string) (Aggregator, error)
	Templater(ctx context.Context, ruleName string) (Templater, error)
	Sender(ctx context.Context, ruleName string) (Sender, error)
}

type Rule interface {
	DeepCopyRule() Rule
}
