package biz

import (
	"context"
	"github.com/aide-family/moon/pkg/conf"
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

	config := GetConfigData()
	receives, ok := config.GetReceivers(msg.Route)
	if !ok {
		return merr.ErrorAlert("receiver not found")
	}

	globalEmailConfig := b.c.GetGlobalEmailConfig()
	// 如果有自定义的邮箱配置， 使用自定义， 否则使用公共邮箱配置
	if !types.IsNil(receives.GetEmailConfig()) {
		globalEmailConfig = receives.GetEmailConfig()
	}

	emailReceives := receives.GetEmails()
	hookReceivers := receives.GetHooks()
	senderList := make([]notify.Notify, 0, len(emailReceives)*4+len(hookReceivers)*4)
	for _, hookItem := range hookReceivers {
		if types.IsNil(hookItem) {
			continue
		}
		hookNotify := &conf.ReceiverHook{
			Type:     hookItem.GetType(),
			Webhook:  hookItem.GetWebhook(),
			Content:  hookItem.GetContent(),
			Template: config.GetTemplates(hookItem.GetTemplate()),
			Secret:   hookItem.GetSecret(),
		}
		newNotify, err := hook.NewNotify(hookNotify)
		if !types.IsNil(err) {
			continue
		}
		senderList = append(senderList, newNotify)
	}

	for _, emailItem := range emailReceives {
		if types.IsNil(emailItem) {
			continue
		}
		emailItemNotify := &conf.ReceiverEmail{
			To:          emailItem.GetTo(),
			Subject:     emailItem.GetSubject(),
			Content:     emailItem.GetContent(),
			Template:    config.GetTemplates(emailItem.GetTemplate()),
			Cc:          emailItem.GetCc(),
			AttachUrl:   emailItem.GetAttachUrl(),
			ContentType: emailItem.GetContentType(),
		}
		senderList = append(senderList, email.New(globalEmailConfig, emailItemNotify))
	}

	// 发送hook告警
	for _, sender := range senderList {
		send := sender
		key := msg.Key(send)
		if !types.TextIsNull(key) {
			ok, err := b.data.GetCacher().SetNX(ctx, key, "ok", 1*time.Hour)
			if err != nil {
				return err
			}
			if !ok {
				return nil
			}
		}

		if err := send.Send(ctx, msgMap); !types.IsNil(err) {
			log.Warnw("send hook error", err, "receiver", send.Type())
			// 删除缓存  加入重试队列
			b.data.GetCacher().Delete(ctx, key)
			return err
		}
	}
	return nil
}
