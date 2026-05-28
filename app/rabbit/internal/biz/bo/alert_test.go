package bo

import (
	"testing"

	"github.com/bwmarrin/snowflake"
)

func TestAlertSubscriptionItemBo_MatchesLabels(t *testing.T) {
	sub := &AlertSubscriptionItemBo{
		Labels: map[string]string{"aa": "aa"},
	}
	if !sub.MatchesLabels(map[string]string{"aa": "aa", "alertname": "up"}) {
		t.Fatal("expected label match")
	}
	if sub.MatchesLabels(map[string]string{"aa": "bb"}) {
		t.Fatal("expected label mismatch")
	}
	if sub.MatchesLabels(map[string]string{}) {
		t.Fatal("expected mismatch when subscription label missing on alert")
	}
}

func TestAlertSubscriptionItemBo_MatchesLabels_exclude(t *testing.T) {
	sub := &AlertSubscriptionItemBo{
		Labels:        map[string]string{"aa": "aa"},
		ExcludeLabels: map[string]string{"severity": "S1"},
	}
	if !sub.MatchesLabels(map[string]string{"aa": "aa", "severity": "S2"}) {
		t.Fatal("expected match when exclude labels do not all match")
	}
	if sub.MatchesLabels(map[string]string{"aa": "aa", "severity": "S1"}) {
		t.Fatal("expected exclude to block match")
	}
}

func TestAlertSubscriptionItemBo_DirectEmailEnabled(t *testing.T) {
	sub := &AlertSubscriptionItemBo{
		DirectMemberEmailConfigUID: snowflake.ID(1),
		Members: []*AlertSubscriptionMemberBo{
			{MemberUID: 100, IsEmail: true},
		},
	}
	if !sub.DirectEmailEnabled() {
		t.Fatal("expected direct email enabled")
	}
	sub.DirectMemberEmailConfigUID = 0
	if sub.DirectEmailEnabled() {
		t.Fatal("expected disabled without email config uid")
	}
	sub.DirectMemberEmailConfigUID = snowflake.ID(1)
	sub.Members[0].IsEmail = false
	if sub.DirectEmailEnabled() {
		t.Fatal("expected disabled without isEmail member")
	}
}
