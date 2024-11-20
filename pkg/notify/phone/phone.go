package phone

import (
	"context"

	"github.com/aide-family/moon/pkg/notify"
	"github.com/go-kratos/kratos/v2/log"
)

// New 创建电话通知
func New() Phone {
	return &p{}
}

type (
	// Phone 电话通知
	Phone interface {
		notify.Notify
		Call(ctx context.Context) error
	}

	p struct{}
)

// Hash 返回通知的唯一标识
func (l *p) Hash() string {
	return l.Type()
}

// Type 返回通知类型
func (l *p) Type() string {
	return "phone"
}

// Send 发送通知
func (l *p) Send(ctx context.Context, msg notify.Msg) error {
	log.Debugw("send phone", "phone", l, "msg", msg)
	return nil
}

// Call 拨打电话
func (l *p) Call(ctx context.Context) error {
	log.Debugw("call phone", "phone", l)
	return nil
}
