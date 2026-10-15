package bo

import (
	"testing"
	"time"
)

func TestDue(t *testing.T) {
	now := time.Date(2026, 10, 15, 12, 0, 0, 0, time.UTC)
	past := now.Add(-time.Minute)
	future := now.Add(time.Minute)
	cases := []struct {
		name string
		rule AlertRule
		want bool
	}{
		{name: "enabled metrics without schedule", rule: AlertRule{Status: "ACTIVE", Type: "METRICS"}, want: true},
		{name: "enabled metrics already due", rule: AlertRule{Status: "ACTIVE", Type: "METRICS", NextEvalAt: &past}, want: true},
		{name: "enabled metrics not yet due", rule: AlertRule{Status: "ACTIVE", Type: "METRICS", NextEvalAt: &future}, want: false},
		{name: "disabled metrics", rule: AlertRule{Status: "DISABLED", Type: "METRICS"}, want: false},
		{name: "enabled logs", rule: AlertRule{Status: "ACTIVE", Type: "LOGS"}, want: false},
		{name: "enabled trace", rule: AlertRule{Status: "ACTIVE", Type: "TRACE"}, want: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.rule.Due(now); got != tc.want {
				t.Fatalf("Due() = %v, want %v", got, tc.want)
			}
		})
	}
}
