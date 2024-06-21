package watch

import (
	"github.com/go-kratos/kratos/v2/log"
)

// 定义接收消息和发送消息的消息队列

type (
	// Queue 消息队列
	Queue interface {
		// Next 获取下一个消息
		Next() (*Message, bool)
		// Push 添加消息
		Push(msg *Message) error
		// Close 关闭队列
		Close() error
		// Len 获取队列长度
		Len() int
		// Clear 清空队列
		Clear()
	}

	// defaultQueue 默认消息队列
	defaultQueue struct {
		queue   chan *Message
		maxSize int
	}
)

func (d *defaultQueue) Next() (*Message, bool) {
	msg, ok := <-d.queue
	return msg, ok
}

func (d *defaultQueue) Push(msg *Message) error {
	log.Debugw("method", "Push", "msg", msg, "data", msg.GetData())
	d.queue <- msg
	return nil
}

func (d *defaultQueue) Close() error {
	close(d.queue)
	return nil
}

func (d *defaultQueue) Len() int {
	return len(d.queue)
}

func (d *defaultQueue) Clear() {
	d.queue = make(chan *Message, d.maxSize)
}

func NewDefaultQueue(size int) Queue {
	return &defaultQueue{
		queue: make(chan *Message, size),
	}
}
