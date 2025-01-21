package watch

import (
	"context"
	"fmt"
	"sync"

	"github.com/aide-family/moon/pkg/plugin/cache"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
)

// NewDefaultStorage 定义运行时存储
func NewDefaultStorage() Storage {
	return &defaultStorage{
		data: make(map[Indexer]*Message),
	}
}

// NewCacheStorage 定义缓存存储器
func NewCacheStorage(cacher cache.ICacher) Storage {
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
		cacher cache.ICacher
	}

	// CacheStorageMsg 缓存存储器消息
	cacheStorageMsg struct {
		Data     Indexer    `json:"data"`
		Topic    vobj.Topic `json:"topic"`
		Retry    int        `json:"retry"`
		RetryMax int        `json:"retry_max"`
	}
)

func (c cacheStorageMsg) String() string {
	bs, _ := types.Marshal(c)
	return string(bs)
}

func newMessage(data string) *Message {
	var msg cacheStorageMsg
	_ = types.Unmarshal([]byte(data), &msg)

	return &Message{
		data:     msg.Data,
		topic:    msg.Topic,
		retry:    msg.Retry,
		retryMax: msg.RetryMax,
	}
}

func (c *cacheStorage) Get(index Indexer) *Message {
	cacheStr, _ := c.cacher.Client().Get(context.Background(), index.Index()).Result()
	return newMessage(cacheStr)
}

func (c *cacheStorage) Put(msg *Message) error {
	if msg.data == nil {
		return nil
	}
	cacheMsg := &cacheStorageMsg{
		Data:     msg.data,
		Topic:    msg.topic,
		Retry:    msg.retry,
		RetryMax: msg.retryMax,
	}
	return c.cacher.Client().Set(context.Background(), cacheMsg.Data.Index(), cacheMsg.String(), 0).Err()
}

func (c *cacheStorage) Clear() {
}

func (c *cacheStorage) Remove(index Indexer) {
	_ = c.cacher.Client().Del(context.Background(), index.Index()).Err()
}

func (c *cacheStorage) Close() error {
	return c.cacher.Close()
}

func (c *cacheStorage) Len() int {
	return 0
}

func (c *cacheStorage) Range(f func(index Indexer, msg *Message) bool) {
	keys, err := c.cacher.Client().Keys(context.Background(), "").Result()
	if err != nil {
		return
	}
	// 遍历缓存存储器，并将缓存消息反序列化为Message
	for _, key := range keys {
		d, err := c.cacher.Client().Get(context.Background(), key).Result()
		if err != nil {
			continue
		}
		msg := newMessage(d)
		if msg.retryMax > 0 && msg.retry >= msg.retryMax {
			continue
		}
		if !f(msg.data, msg) {
			break
		}
	}
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
