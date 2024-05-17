package biz

import (
	"context"
	"encoding/json"

	"github.com/aide-cloud/moon/api/merr"
	"github.com/aide-cloud/moon/cmd/server/rabbit/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/rabbit/internal/rabbitconf"
	"github.com/aide-cloud/moon/pkg/notify"
	"github.com/aide-cloud/moon/pkg/notify/email"
	"github.com/aide-cloud/moon/pkg/notify/hook"
	"github.com/aide-cloud/moon/pkg/types"
)

func NewMsgBiz(c *rabbitconf.Bootstrap) *MsgBiz {
	return &MsgBiz{c: c}
}

type MsgBiz struct {
	c *rabbitconf.Bootstrap
}

func (b *MsgBiz) SendMsg(ctx context.Context, msg *bo.SendMsgParams) error {
	var msgMap notify.Msg
	if err := json.Unmarshal(msg.Data, &msgMap); err != nil {
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

		if err := email.New(globalEmailConfig, emailItem).Send(ctx, msgMap); err != nil {
			return err
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
		if err != nil {
			continue
		}
		if err := newNotify.Send(ctx, msgMap); err != nil {
			continue
		}
	}
	return nil
}
