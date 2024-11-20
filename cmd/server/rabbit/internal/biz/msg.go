package biz

import (
	"context"
	"sync"
	"time"

	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/data"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/conf"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/notify/email"
	"github.com/aide-family/moon/pkg/notify/hook"
	"github.com/aide-family/moon/pkg/util/after"
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

// SendMsg 发送消息  TODO 标记后续改造发送通知
func (b *MsgBiz) SendMsg(ctx context.Context, msg *bo.SendMsgParams) error {
	if types.TextIsNull(msg.Route) {
		return nil
	}

	config := GetConfigData()
	receives, ok := config.GetReceivers(msg.Route)
	if !ok {
		return merr.ErrorAlert("%s receiver not found", msg.Route)
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

	// 异步发送消息
	b.sendMsg(ctx, msg, senderList...)
	return nil
}

func (b *MsgBiz) sendMsg(ctx context.Context, msg *bo.SendMsgParams, sends ...notify.Notify) {
	if types.IsNil(msg) || len(sends) == 0 {
		return
	}

	var msgMap notify.Msg
	if err := types.Unmarshal(msg.Data, &msgMap); !types.IsNil(err) {
		return
	}

	if len(msgMap) == 0 {
		return
	}

	var wg sync.WaitGroup
	// 发送hook告警
	for _, sender := range sends {
		if sender == nil {
			continue
		}
		wg.Add(1)
		go func(send notify.Notify) {
			defer after.RecoverX()
			defer wg.Done()
			key := msg.Key(send)
			if types.TextIsNull(key) {
				return
			}

			nxOK, err := b.data.GetCacher().SetNX(ctx, key, "ok", 1*time.Hour)
			if err != nil {
				log.Warnw("method", "set cache error", "err", err)
				return
			}
			if !nxOK {
				return
			}

			if err := send.Send(ctx, msgMap); !types.IsNil(err) {
				log.Warnw("method", "send hook error", "err", err, "receiver", send.Type())
				// 删除缓存  加入重试队列
				_ = b.data.GetCacher().Delete(ctx, key)
			}
		}(sender)
	}
	wg.Wait()
}
