package bo

import (
	"testing"

	"github.com/aide-family/magicbox/enum"
)

func TestIsWebhookMessageType(t *testing.T) {
	cases := []struct {
		messageType enum.MessageType
		want        bool
	}{
		{enum.MessageType_EMAIL, false},
		{enum.MessageType_SMS_ALICLOUD, false},
		{enum.MessageType_WEBHOOK_OTHER, true},
		{enum.MessageType_WEBHOOK_DINGTALK, true},
		{enum.MessageType_WEBHOOK_WECHAT, true},
		{enum.MessageType_WEBHOOK_FEISHU, true},
		{enum.MessageType(2999), true},
		{enum.MessageType(3000), false},
	}
	for _, tc := range cases {
		if got := IsWebhookMessageType(tc.messageType); got != tc.want {
			t.Fatalf("IsWebhookMessageType(%v) = %v, want %v", tc.messageType, got, tc.want)
		}
	}
}
