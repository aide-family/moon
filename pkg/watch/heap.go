package watch

import (
	"context"
	"fmt"
	"sync"

	"github.com/aide-family/moon/pkg/util/conn"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewDefaultStorage 定义运行时存储
func NewDefaultStorage() Storage {
	return &defaultStorage{
		data: make(map[Indexer]*Message),
	}
}

// NewCacheStorage 定义缓存存储器
func NewCacheStorage(cacher conn.Cache) Storage {
	return &cacheStorage{
		cacher: cacher,
	}
}

type (
	// Indexer 索引器
	Indexer interface {
		fmt.Stringer
		// Index 索引生成器
		Index() string
	}

	// Storage 存储器
	Storage interface {
		// Get 获取消息
		Get(index Indexer) *Message

		// Put 放入消息
		Put(msg *Message) error

		// Clear 清空消息
		Clear()

		// Remove 移除消息
		Remove(index Indexer)

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

	// cacheStorage 缓存存储器
	cacheStorage struct {
		cacher conn.Cache
	}

	// CacheStorageMsg 缓存存储器消息
	CacheStorageMsg string
)

func (c CacheStorageMsg) String() string {
	return string(c)
}

// Index 实现Indexer接口
func (c CacheStorageMsg) Index() string {
	return string(c)
}

// Unmarshal 反序列化
func (c CacheStorageMsg) Unmarshal(v any) error {
	return types.Unmarshal([]byte(c), v)
}

func (c *cacheStorage) Get(index Indexer) *Message {
	var msg Message
	cacheStr, err := c.cacher.Get(context.Background(), index.Index())
	if err == nil {
		msg.data = CacheStorageMsg(cacheStr)
	}

	return &msg
}

func (c *cacheStorage) Put(msg *Message) error {
	return c.cacher.Set(context.Background(), msg.data.Index(), msg.data.String(), 0)
}

func (c *cacheStorage) Clear() {
}

func (c *cacheStorage) Remove(index Indexer) {
	c.cacher.Delete(context.Background(), index.Index())
}

func (c *cacheStorage) Close() error {
	return c.cacher.Close()
}

func (c *cacheStorage) Len() int {
	return 0
}

func (c *cacheStorage) Range(f func(index Indexer, msg *Message) bool) {
}

func (d *defaultStorage) Clear() {
	d.lock.Lock()
	defer d.lock.Unlock()
	d.data = make(map[Indexer]*Message)
}

func (d *defaultStorage) Remove(index Indexer) {
	d.lock.Lock()
	defer d.lock.Unlock()
	delete(d.data, index)
}

func (d *defaultStorage) Range(f func(index Indexer, msg *Message) bool) {
	d.lock.Lock()
	copyMap := make(map[Indexer]*Message)
	for k, v := range d.data {
		if !types.IsNil(v) {
			copyMap[k] = v
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
