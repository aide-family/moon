package watch

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
)
