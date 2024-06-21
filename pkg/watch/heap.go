package watch

import (
	"sync"

	"github.com/aide-family/moon/pkg/util/types"
)

// 定义运行时存储

func NewDefaultStorage() Storage {
	return &defaultStorage{
		data: make(map[Indexer]*Message),
	}
}

type (
	// Indexer 索引器
	Indexer interface {
		// Index 索引生成器
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

		// Len 长度
		Len() int

		// Range 遍历
		//  f返回值为bool类型，如果返回false，则停止range
		Range(f func(index Indexer, msg *Message) bool)
	}

	// defaultStorage 默认存储器
	defaultStorage struct {
		lock sync.Mutex
		data map[Indexer]*Message
	}
)

func (d *defaultStorage) Range(f func(index Indexer, msg *Message) bool) {
	d.lock.Lock()
	copyMap := make(map[Indexer]*Message)
	for k, v := range d.data {
		if !types.IsNil(v) {
			msg := *v
			copyMap[k] = &msg
		}
	}
	defer d.lock.Unlock()
	for k, v := range copyMap {
		if !f(k, v) {
			break
		}
	}
}

func (d *defaultStorage) Len() int {
	d.lock.Lock()
	defer d.lock.Unlock()
	return len(d.data)
}

func (d *defaultStorage) Get(index Indexer) *Message {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.data[index]
}

func (d *defaultStorage) Put(msg *Message) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data[msg.GetData()] = msg
	return nil
}

func (d *defaultStorage) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data = nil
	return nil
}
