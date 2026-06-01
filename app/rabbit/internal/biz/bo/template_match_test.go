package bo

import (
	"testing"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"
)

func TestMatchWebhookTemplate(t *testing.T) {
	feishuTemplateUID := snowflake.ParseInt64(100)
	templates := []*TemplateItemBo{
		nil,
		{
			UID:         feishuTemplateUID,
			MessageType: enum.MessageType_WEBHOOK_FEISHU,
			Status:      enum.GlobalStatus_ENABLED,
		},
		{
			UID:         snowflake.ParseInt64(200),
			MessageType: enum.MessageType_WEBHOOK_DINGTALK,
			Status:      enum.GlobalStatus_ENABLED,
		},
		{
			UID:         snowflake.ParseInt64(300),
			MessageType: enum.MessageType_WEBHOOK_FEISHU,
			Status:      enum.GlobalStatus_DISABLED,
		},
	}

	got := MatchWebhookTemplate(templates, enum.WebhookAPP_FEISHU)
	if got != feishuTemplateUID {
		t.Fatalf("MatchWebhookTemplate() = %v, want %v", got, feishuTemplateUID)
	}
	if got := MatchWebhookTemplate(nil, enum.WebhookAPP_FEISHU); got != 0 {
		t.Fatalf("MatchWebhookTemplate(nil) = %v, want 0", got)
	}
}
