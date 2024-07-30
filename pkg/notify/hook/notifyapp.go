package hook

import (
	"github.com/aide-family/moon/api"
	"github.com/aide-family/moon/pkg/notify"
	"github.com/go-kratos/kratos/v2/errors"
)

type (
	// Notify 通知通用接口
	Notify interface {
		notify.Notify
	}
)

// NewNotify 创建通知
func NewNotify(receiverHook any) (Notify, error) {
	switch config := receiverHook.(type) {
	case *api.ReceiverHookDingTalk:
		return NewDingTalk(config), nil
	case *api.ReceiverHookWechatWork:
		return NewWechat(config), nil
	case *api.ReceiverHookFeiShu:
		return NewFeiShu(config), nil
	case *api.ReceiverHookOther:
		return NewOther(config), nil
	default:
		return nil, errors.New(404, "notify.hook.NewNotify", "notify app not support")
	}
}
