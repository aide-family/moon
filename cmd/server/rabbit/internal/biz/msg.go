package biz

import (
	"context"
	"time"

	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/data"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/notify/email"
	"github.com/aide-family/moon/pkg/notify/hook"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
	"golang.org/x/sync/errgroup"
)

// NewMsgBiz 创建消息业务
func NewMsgBiz(c *rabbitconf.Bootstrap, data *data.Data) *MsgBiz {
	return &MsgBiz{c: c, data: data}
}

// MsgBiz 消息业务
type MsgBiz struct {
	c    *rabbitconf.Bootstrap
	data *data.Data
}

// SendMsg 发送消息
func (b *MsgBiz) SendMsg(ctx context.Context, msg *bo.SendMsgParams) error {
	var msgMap notify.Msg
	if err := types.Unmarshal(msg.Data, &msgMap); !types.IsNil(err) {
		return err
	}

	hookData := GetConfigData()
	receives := hookData.GetReceivers()[msg.Route]
	if types.IsNil(receives) {
		return merr.ErrorAlert("receiver not found")
	}

	globalEmailConfig := b.c.GetGlobalEmailConfig()
	// 如果有自定义的邮箱配置， 使用自定义， 否则使用公共邮箱配置
	if !types.IsNil(receives.GetEmailConfig()) {
		globalEmailConfig = receives.GetEmailConfig()
	}

	tempMap := b.c.GetTemplates()
	emailReceives := receives.GetEmails()
	hookReceivers := receives.GetHooks()
	hookList := make([]hook.Config, 0, len(hookReceivers)*4)
	for _, hookItem := range hookReceivers {
		if types.IsNil(hookItem) {
			continue
		}
		hookItem.Template = tempMap[hookItem.GetTemplate()]
		hookList = append(hookList, hookItem)
	}

	eg := new(errgroup.Group)
	for _, emailItem := range emailReceives {
		if types.IsNil(emailItem) {
			continue
		}
		tempKey := emailItem.GetTemplate()
		if temp, ok := tempMap[tempKey]; ok {
			emailItem.Template = temp
		}

		eg.Go(func() error {
			if err := email.New(globalEmailConfig, emailItem).Send(ctx, msgMap); !types.IsNil(err) {
				log.Warnw("send email error", err, "receiver", emailItem)
				return err
			}
			return nil
		})
	}

	// 发送hook告警
	for _, hookItem := range hookList {
		newNotify, err := hook.NewNotify(hookItem)
		if !types.IsNil(err) {
			continue
		}
		eg.Go(func() error {
			key := msg.Key(newNotify.Type())
			if !types.TextIsNull(key) {
				ok, err := b.data.GetCacher().SetNX(ctx, msg.Key(newNotify.Type()), "ok", 1*time.Hour)
				if err != nil {
					return err
				}
				if !ok {
					return nil
				}
			}

			if err := newNotify.Send(ctx, msgMap); !types.IsNil(err) {
				log.Warnw("send hook error", err, "receiver", hookItem)
				return err
			}
			return nil
		})
	}
	return eg.Wait()
}
