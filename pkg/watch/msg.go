package watch

import (
	"context"
	"sync"

	"github.com/aide-family/moon/pkg/vobj"
)

// 定义原始消息格式和传输消息格式

const defaultRetryMax = 0

// NewMessage 创建消息
func NewMessage(data Indexer, topic vobj.Topic, opts ...MessageOption) *Message {
	m := &Message{
		data:      data,
		topic:     topic,
		schema:    NewEmptySchemer(),
		retry:     0,
		retryMax:  defaultRetryMax,
		handleCtx: context.Background(),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

type (
	// Message watch 消息结构体
	Message struct {
		lock sync.Mutex

		// 传输的消息内容， 由用户自定义
		data Indexer

		// 消息类型， 如需要增加新的类型，去vobj包增加
		topic vobj.Topic

		// 注册编码器
		schema Schemer

		// 重试次数
		retry int

		// 最大消息重试次数
		retryMax int

		// 是否已经处理过
		handleCtx context.Context
	}

	MessageOption func(m *Message)
)

// GetData 获取消息内容
func (m *Message) GetData() Indexer {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.data
}

// GetTopic 获取消息类型
func (m *Message) GetTopic() vobj.Topic {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.topic
}

// GetSchema 获取消息编码器
func (m *Message) GetSchema() Schemer {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.schema
}

// GetRetry 获取消息重试次数
func (m *Message) GetRetry() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.retry
}

// RetryInc 重试次数+1
func (m *Message) RetryInc() {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.retry++
}

// GetRetryMax 获取消息最大重试次数
func (m *Message) GetRetryMax() int {
	m.lock.Lock()
	defer m.lock.Unlock()
	return m.retryMax
}

// WithHandledPath 设置消息处理过的路径
func (m *Message) WithHandledPath(index int, handle HandleFun) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.handleCtx = context.WithValue(m.handleCtx, index, handle)
}

// IsHandled 判断消息是否已经处理过
func (m *Message) IsHandled(index int) bool {
	m.lock.Lock()
	defer m.lock.Unlock()
	v := m.handleCtx.Value(index)
	if v == nil {
		return false
	}
	_, ok := v.(HandleFun)
	return ok
}

// WithMessageSchema 设置消息编码器
func WithMessageSchema(schema Schemer) MessageOption {
	return func(m *Message) {
		m.schema = schema
	}
}

// WithMessageRetryMax 设置消息最大重试次数
func WithMessageRetryMax(retryMax int) MessageOption {
	return func(m *Message) {
		m.retryMax = retryMax
	}
}
