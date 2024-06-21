package watch

// 定义运行时存储

type (
	// Indexer 索引器
	Indexer interface {
		Index() string
	}

	// Storage 存储器
	Storage interface {
		// Get 获取消息
		Get(index Indexer) *Message
		// Put 放入消息
		Put(msg *Message) error
		// Close 关闭存储器
		Close() error
	}
)
