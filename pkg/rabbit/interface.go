package rabbit

import (
	"context"
	"io"
)

// Plugin 仅仅自定义一个Name的接口
// 整体来看，每个模块都是一个插件，它们都是可插拔的
type Plugin interface {
	Name() string
}

// Receiver 用来接收来自外部消息
type Receiver interface {
	Plugin
	Receive() (<-chan *Message, error)
}

// Suppressor 负责按照规则对消息进行抑制。
// Suppressor 需要在 Message 被提案时告知调用者，该消息的发送 Propose 是否通过。
// 如果为未通过，则该条消息将会被抑制:
// + 聚合: 按照规则将消息收集起来等待合适的时机一起发送
// + 丢弃: 抛弃掉该条消息
type Suppressor interface {
	Plugin
	Propose(index string) bool
	Cancel(index string)
	Finish(index string)
}

// Templater 负责对消息进行模版的解析。
type Templater interface {
	Parse(in any, out io.Writer) error
}

// Sender 负责发送消息，它不需要关心什么时候发送，也不需要关心消息的内容是什么。
// 它只负责在接收到消息时，将它按照已经确定的方式发送出去。
type Sender interface {
	Plugin
	Send(context context.Context, content []byte, secret []byte) error
}

// SecretProvider Sender 发送消息时，有的需要 Secret 才能够将送达。
// 然而 Sender 并不需要关心 Secret 如何正确生成，它只需调用 Provider 接口，获取到 Secret 使用即可。
// 之所以这样设计，是因为:
// + 不同的 Sender 需要的 Secret 的数据结构不同
// + 相同的 Sender 不同的接收者需要的 Secret 不同
type SecretProvider interface {
	// Provider 负责在被调用时提供正确的密钥，否则返回错误
	Provider(in []byte, out any) error
}

type ConfigGetter interface {
	GetTemplater(context context.Context, id int64) (Templater, error)
	GetSecret(context context.Context, id int64) ([]byte, error)
	GetSuppressorByTemplate(context context.Context, id int64) (Suppressor, error)
	GetSenderByTemplate(context context.Context, id int64) (Sender, error)
}

type SenderGetter interface {
	Get(context context.Context, name string) (Sender, error)
}

type TemplaterGetter interface {
	Get(context context.Context, id int64) (Templater, error)
}

type SuppressorGetter interface {
	Get(context context.Context, id int64) (Suppressor, error)
}
