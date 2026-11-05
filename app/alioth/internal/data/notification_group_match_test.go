package data

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/internal/data/ent"
)

func TestGroupMatchesUser(t *testing.T) {
	cases := []struct {
		name      string
		group     *ent.NotificationGroup
		memberUID string
		tokens    []string
		want      bool
	}{
		{
			name:      "旧版成员 UID 命中",
			group:     &ent.NotificationGroup{MemberUids: []string{"member-1", "member-2"}},
			memberUID: "member-2",
			want:      true,
		},
		{
			name:      "邮箱成员 UID 命中",
			group:     &ent.NotificationGroup{EmailMemberUids: []string{"member-1"}},
			memberUID: "member-1",
			want:      true,
		},
		{
			name:      "短信成员 UID 命中",
			group:     &ent.NotificationGroup{SmsMemberUids: []string{"member-1"}},
			memberUID: "member-1",
			want:      true,
		},
		{
			name:      "旧版手填邮箱命中",
			group:     &ent.NotificationGroup{Receivers: "alice@example.com,bob@example.com"},
			memberUID: "member-1",
			tokens:    []string{"bob@example.com"},
			want:      true,
		},
		{
			name:   "手填邮箱命中",
			group:  &ent.NotificationGroup{EmailReceivers: "alice@example.com, bob@example.com"},
			tokens: []string{"  BOB@example.com  "},
			want:   true,
		},
		{
			name:   "手填手机号命中",
			group:  &ent.NotificationGroup{SmsReceivers: "13800000000"},
			tokens: []string{"13900000000", "13800000000"},
			want:   true,
		},
		{
			name:      "成员不在任何接收人里",
			group:     &ent.NotificationGroup{EmailMemberUids: []string{"member-2"}, EmailReceivers: "alice@example.com"},
			memberUID: "member-1",
			tokens:    []string{"bob@example.com"},
			want:      false,
		},
		{
			name:   "只按完整接收方匹配，不做前缀",
			group:  &ent.NotificationGroup{EmailReceivers: "alice@example.com.cn"},
			tokens: []string{"alice@example.com"},
			want:   false,
		},
		{
			name:      "空成员和空接收方",
			group:     &ent.NotificationGroup{},
			memberUID: "",
			tokens:    []string{"", "  "},
			want:      false,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := groupMatchesUser(tc.group, tc.memberUID, tc.tokens); got != tc.want {
				t.Fatalf("groupMatchesUser = %v, want %v", got, tc.want)
			}
		})
	}
}
