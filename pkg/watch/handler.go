package watch

import (
	"context"
	"fmt"
	"sync"

	"github.com/aide-family/moon/pkg/vobj"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewDefaultHandler(opts ...DefaultHandlerOption) Handler {
	d := &defaultHandler{
		topicHandleMap: make(map[vobj.Topic][]HandleFun),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

type (
	// Handler 消息处理
	Handler interface {
		// Handle 处理消息
		//
		// 	ctx 上下文
		// 	msg 消息
		Handle(ctx context.Context, msg *Message) error
	}

	HandleFun func(ctx context.Context, msg *Message) error

	// defaultHandler 默认消息处理
	defaultHandler struct {
		lock           sync.Mutex
		topicHandleMap map[vobj.Topic][]HandleFun
	}

	DefaultHandlerOption func(d *defaultHandler)
)

func (d *defaultHandler) Handle(ctx context.Context, msg *Message) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	// 获取处理器
	handles, ok := d.topicHandleMap[msg.GetTopic()]
	if !ok {
		return status.Error(codes.Unimplemented, fmt.Sprintf("%s topic not found handle", msg.GetTopic()))
	}

	// 调用处理器处理msg
	for index, handle := range handles {
		// 消息已经被此handle处理过
		if msg.IsHandled(index) {
			continue
		}
		if err := handle(ctx, msg); err != nil {
			return err
		}
		// 标记消息处理状态， 避免被重播
		msg.WithHandledPath(index, handle)
	}
	return nil
}

func WithDefaultHandlerTopicHandle(topic vobj.Topic, handles ...HandleFun) DefaultHandlerOption {
	return func(d *defaultHandler) {
		d.lock.Lock()
		defer d.lock.Unlock()
		d.topicHandleMap[topic] = handles
	}
}
