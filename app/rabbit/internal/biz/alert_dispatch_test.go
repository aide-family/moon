package biz

import (
	"testing"

	"github.com/bwmarrin/snowflake"

	"github.com/aide-family/rabbit/internal/biz/bo"
)

func TestUniqueInt64IDs(t *testing.T) {
	got := uniqueInt64IDs([]int64{1, 2, 1, 0, 3, 2})
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Fatalf("unexpected result: %v", got)
	}
}

func TestMergeAlertEmailDispatchPlans(t *testing.T) {
	plans := make(map[int64]*alertEmailDispatchPlan)
	cfg := snowflake.ParseInt64(10)
	mergeAlertEmailDispatchPlans(plans, cfg, []string{"a@x.com"})
	mergeAlertEmailDispatchPlans(plans, cfg, []string{"b@x.com", "a@x.com"})
	plan := plans[cfg.Int64()]
	if plan == nil || len(plan.to) != 2 {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if plan.to[0] != "a@x.com" || plan.to[1] != "b@x.com" {
		t.Fatalf("unexpected recipients: %v", plan.to)
	}
}

func TestRecipientGroupEmailRecipients(t *testing.T) {
	group := &bo.RecipientGroupItemBo{
		Members: []*bo.NotificationMemberBo{
			{IsEmail: true, MemberEmail: "a@x.com"},
			{IsEmail: true, MemberEmail: "a@x.com"},
			{IsEmail: false, MemberEmail: "skip@x.com"},
		},
	}
	to := recipientGroupEmailRecipients(group)
	if len(to) != 1 || to[0] != "a@x.com" {
		t.Fatalf("unexpected recipients: %v", to)
	}
}

func TestSubscriptionNotificationRouteKeys(t *testing.T) {
	subUID := snowflake.ParseInt64(1)
	cfgUID := snowflake.ParseInt64(2)
	if got := subscriptionEmailRouteKey(subUID, cfgUID); got != "sub:1:email:2" {
		t.Fatalf("email route key = %q", got)
	}
	if got := subscriptionWebhookRouteKey(subUID, cfgUID); got != "sub:1:webhook:2" {
		t.Fatalf("webhook route key = %q", got)
	}
}
