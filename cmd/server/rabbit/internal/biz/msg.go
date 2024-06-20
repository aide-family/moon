package biz

import (
	"context"
	"encoding/json"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-family/moon/pkg/notify"
	"github.com/aide-family/moon/pkg/notify/email"
	"github.com/aide-family/moon/pkg/notify/hook"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/log"
)

func NewMsgBiz(c *rabbitconf.Bootstrap) *MsgBiz {
	return &MsgBiz{c: c}
}

type MsgBiz struct {
	c *rabbitconf.Bootstrap
}

func (b *MsgBiz) SendMsg(ctx context.Context, msg *bo.SendMsgParams) error {
	var msgMap notify.Msg
	if err := json.Unmarshal(msg.Data, &msgMap); !types.IsNil(err) {
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
	for _, emailItem := range emailReceives {
		if types.IsNil(emailItem) {
			continue
		}
		tempKey := emailItem.GetTemplate()
		if temp, ok := tempMap[tempKey]; ok {
			emailItem.Template = temp
		}

		if err := email.New(globalEmailConfig, emailItem).Send(ctx, msgMap); !types.IsNil(err) {
			log.Warnw("send email error", err, "receiver", emailItem)
			continue
		}
	}

	hookReceivers := receives.GetHooks()
	hookList := make([]any, 0, len(hookReceivers)*4)
	for _, hookItem := range hookReceivers {
		if types.IsNil(hookItem) {
			continue
		}
		if !types.IsNil(hookItem.GetDingTalk()) {
			dingTalkItem := hookItem.GetDingTalk()
			dingTalkItem.Template = tempMap[dingTalkItem.GetTemplate()]
			hookList = append(hookList, dingTalkItem)
		}
		if !types.IsNil(hookItem.GetWechatWork()) {
			wechatWorkItem := hookItem.GetWechatWork()
			wechatWorkItem.Template = tempMap[wechatWorkItem.GetTemplate()]
			hookList = append(hookList, wechatWorkItem)
		}
		if !types.IsNil(hookItem.GetFeiShu()) {
			feishuItem := hookItem.GetFeiShu()
			feishuItem.Template = tempMap[feishuItem.GetTemplate()]
			hookList = append(hookList, feishuItem)
		}
		if !types.IsNil(hookItem.GetOther()) {
			otherItem := hookItem.GetOther()
			otherItem.Template = tempMap[otherItem.GetTemplate()]
			hookList = append(hookList, otherItem)
		}
	}
	// 发送hook告警
	for _, hookItem := range hookList {
		newNotify, err := hook.NewNotify(hookItem)
		if !types.IsNil(err) {
			continue
		}
		if err := newNotify.Send(ctx, msgMap); !types.IsNil(err) {
			log.Warnw("send hook error", err, "receiver", hookItem)
			continue
		}
	}
	return nil
}
