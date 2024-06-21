package watch

import (
	"context"
	"fmt"
	"sync"

	"github.com/aide-family/moon/pkg/vobj"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type (
	// Handler 消息处理
	Handler interface {
		Handle(ctx context.Context, msg *Message, storage Storage) error
	}

	HandleFun func(ctx context.Context, msg *Message) error

	// defaultHandler 默认消息处理
	defaultHandler struct {
		lock           sync.Mutex
		topicHandleMap map[vobj.Topic]HandleFun
	}

	DefaultHandlerOption func(d *defaultHandler)
)

func (d *defaultHandler) Handle(ctx context.Context, msg *Message, storage Storage) error {
	d.lock.Lock()
	defer d.lock.Unlock()
	storage.Put(msg)
	handle, ok := d.topicHandleMap[msg.GetTopic()]
	if !ok {
		return status.Error(codes.Unimplemented, fmt.Sprintf("%s topic not found handle", msg.GetTopic()))
	}

	return handle(ctx, msg)
}

func NewDefaultHandler(opts ...DefaultHandlerOption) Handler {
	d := &defaultHandler{
		topicHandleMap: make(map[vobj.Topic]HandleFun),
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

func WithDefaultHandlerTopicHandle(topic vobj.Topic, handle HandleFun) DefaultHandlerOption {
	return func(d *defaultHandler) {
		d.topicHandleMap[topic] = handle
	}
}
