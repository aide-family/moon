package mq

import "github.com/aide-family/moon/pkg/util/types"

// IMQ mq接口
type IMQ interface {
	// Send 发送消息
	Send(topic string, data []byte) error

	// Receive 接收消息 返回一个接收通道
	Receive(topic string) <-chan *Msg

	// RemoveReceiver 移除某个topic的接收通道
	RemoveReceiver(topic string)

	// Close 关闭连接
	Close()
}

var _ IMQ = (*mockMQ)(nil)

// NewMockMQ 创建一个mock的mq
func NewMockMQ() IMQ {
	return &mockMQ{
		q: make(map[string]chan *Msg, 100),
	}
}

type (
	mockMQ struct {
		q map[string]chan *Msg
	}

	// Msg 消息结构
	Msg struct {
		Data      []byte
		Topic     []byte
		Timestamp *types.Time
	}
)

func (m *mockMQ) RemoveReceiver(topic string) {
	ch, ok := m.q[topic]
	if !ok {
		return
	}
	close(ch)
	delete(m.q, topic)
}

func (m *mockMQ) Receive(topic string) <-chan *Msg {
	if _, ok := m.q[topic]; !ok {
		m.q[topic] = make(chan *Msg, 100)
	}
	return m.q[topic]
}

func (m *mockMQ) Send(topic string, data []byte) error {
	if m.q == nil {
		m.q = make(map[string]chan *Msg, 100)
	}
	topicCh, ok := m.q[topic]
	if !ok {
		topicCh = make(chan *Msg, 100)
		m.q[topic] = topicCh
	}
	topicCh <- &Msg{Data: data, Topic: []byte(topic)}
	return nil
}

func (m *mockMQ) Close() {
	for _, msgs := range m.q {
		close(msgs)
	}
	m.q = nil
}
