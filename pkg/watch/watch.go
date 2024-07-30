package watch

import (
	"context"
	"time"

	"github.com/aide-family/moon/pkg/util/after"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/transport"
)

var _ transport.Server = (*Watcher)(nil)

const watcherTimeout = time.Second * 10

// NewWatcher 创建监听器
func NewWatcher(name string, opts ...WatcherOption) *Watcher {
	w := &Watcher{
		name:    name,
		stopCh:  make(chan struct{}),
		timeout: watcherTimeout,
	}
	for _, opt := range opts {
		opt(w)
	}
	return w
}

type (
	// Watcher 监听器
	Watcher struct {
		// 服务名称
		name string
		// 停止监听的通道
		stopCh chan struct{}
		// 存储器
		storage Storage
		// 消息队列
		queue Queue
		// 消息处理器
		handler Handler
		// 超时时间
		timeout time.Duration
	}

	// WatcherOption 监听器配置
	WatcherOption func(w *Watcher)
)

// GetStorage 获取存储器
func (w *Watcher) GetStorage() Storage {
	return w.storage
}

// GetQueue 获取消息队列
func (w *Watcher) GetQueue() Queue {
	return w.queue
}

// GetHandler 获取消息处理器
func (w *Watcher) GetHandler() Handler {
	return w.handler
}

// Start 启动监听
func (w *Watcher) Start(_ context.Context) error {
	go func() {
		defer after.RecoverX()
		for {
			select {
			case <-w.stopCh:
				log.Infow("method", "stop watcher")
				w.clear()
				return
			case msg, ok := <-w.queue.Next():
				if !ok {
					continue
				}
				w.reader(msg)
			}
		}
	}()
	log.Infof("[Watcher] %s server started", w.name)
	return nil
}

// Stop 停止监听
func (w *Watcher) Stop(_ context.Context) error {
	defer log.Infof("[Watcher] %s server stoped", w.name)
	w.stopCh <- struct{}{}
	return nil
}

// clear 清理资源
func (w *Watcher) clear() {
	if !types.IsNil(w.queue) {
		if err := w.queue.Close(); err != nil {
			log.Errorw("method", "close queue error", "error", err)
		}
	}

	if !types.IsNil(w.storage) {
		if err := w.storage.Close(); err != nil {
			log.Errorw("method", "close storage error", "error", err)
		}
	}

	close(w.stopCh)
	log.Infow("method", "clear resources", "res", "done")
}

// retry 重试
func (w *Watcher) retry(msg *Message) {
	if msg.GetRetry() >= msg.GetRetryMax() {
		// 重试次数超过最大次数不再重试
		return
	}
	// 消息重试次数+1
	msg.RetryInc()
	if err := w.queue.Push(msg); err != nil {
		log.Errorw("method", "push message to queue error", "error", err)
	}
}

func (w *Watcher) reader(msg *Message) {
	log.Debugw("method", "reader message", "msg", msg)
	if !types.IsNil(w.handler) {
		// 递交消息给处理器，由处理器决定消息去留， 如果失败，会进入重试逻辑
		ctx, cancel := context.WithTimeout(context.Background(), w.timeout)
		defer cancel()
		if err := w.handler.Handle(ctx, msg); err != nil {
			log.Errorw("method", "handle message error", "error", err)
			w.retry(msg)
			return
		}
	}

	if !types.IsNil(w.storage) {
		// 存储消息
		if err := w.storage.Put(msg); err != nil {
			log.Errorw("method", "put message to storage error", "error", err)
			w.retry(msg)
			return
		}
	}
}

// WithWatcherStorage 设置存储器
func WithWatcherStorage(storage Storage) WatcherOption {
	return func(w *Watcher) {
		w.storage = storage
	}
}

// WithWatcherQueue 设置消息队列
func WithWatcherQueue(queue Queue) WatcherOption {
	return func(w *Watcher) {
		w.queue = queue
	}
}

// WithWatcherHandler 设置消息处理器
func WithWatcherHandler(handler Handler) WatcherOption {
	return func(w *Watcher) {
		w.handler = handler
	}
}

// WithWatcherTimeout 设置超时时间
func WithWatcherTimeout(timeout time.Duration) WatcherOption {
	return func(w *Watcher) {
		w.timeout = timeout
	}
}
