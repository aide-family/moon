package hook

import (
	"strings"

	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/go-kratos/kratos/v2/errors"
)

type (
	// Notify 通知通用接口
	Notify interface {
		notify.Notify
	}

	// Config 通知配置
	Config interface {
		GetWebhook() string
		GetSecret() string
		GetContent() string
		GetTemplate() string
		GetType() string
	}
)

// NewNotify 创建通知
func NewNotify(config Config) (Notify, error) {
	switch strings.ToLower(config.GetType()) {
	case vobj.HookAPPDingTalk.EnUSString():
		return NewDingTalk(config), nil
	case vobj.HookAPPWeChat.EnUSString():
		return NewWechat(config), nil
	case vobj.HookAPPFeiShu.EnUSString():
		return NewFeiShu(config), nil
	case vobj.HookAPPWebHook.EnUSString():
		return NewOther(config), nil
	default:
		return nil, errors.New(404, "notify.hook.NewNotify", "notify app not support")
	}
}
