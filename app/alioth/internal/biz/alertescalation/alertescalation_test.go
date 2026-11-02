package alertescalation

import (
	"strings"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// escalationTestPolicy 构造一条字段齐全、可直接交给 normalizePolicy 的策略。
func escalationTestPolicy() bo.AlertEscalationPolicy {
	return bo.AlertEscalationPolicy{
		Name:        "核心告警升级",
		Description: "5 分钟后升级到值班组",
		Labels:      `{"env":"prod"}`,
		LevelUIDs:   []string{"level-p1", "level-p1", ""},
		Steps: []bo.AlertEscalationStep{
			{DelayMinutes: 5, NotificationGroupUIDs: []string{"group-1"}},
			{DelayMinutes: 15, NotificationGroupUIDs: []string{"group-1", "group-2", "group-1"}},
		},
	}
}

// requireInvalid 断言错误是客户端可读的非法参数错误。
func requireInvalid(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected invalid error, got nil")
	}
	if !merr.IsInvalid(err) {
		t.Fatalf("expected invalid error, got %v", err)
	}
}

func TestNormalizePolicy(t *testing.T) {
	policy := escalationTestPolicy()
	if err := normalizePolicy(&policy); err != nil {
		t.Fatalf("normalize valid policy: %v", err)
	}
	if policy.Name != "核心告警升级" || policy.Description != "5 分钟后升级到值班组" {
		t.Fatalf("name or description not normalized: %+v", policy)
	}
	if policy.Labels != `{"env":"prod"}` {
		t.Fatalf("labels = %q", policy.Labels)
	}
	if len(policy.LevelUIDs) != 1 || policy.LevelUIDs[0] != "level-p1" {
		t.Fatalf("level uids not deduplicated: %v", policy.LevelUIDs)
	}
	if got := policy.Steps[1].NotificationGroupUIDs; len(got) != 2 || got[0] != "group-1" || got[1] != "group-2" {
		t.Fatalf("notification groups not deduplicated: %v", got)
	}
}

func TestNormalizePolicyDefaultsEmptyLabels(t *testing.T) {
	policy := escalationTestPolicy()
	policy.Labels = "  "
	if err := normalizePolicy(&policy); err != nil {
		t.Fatalf("normalize empty labels: %v", err)
	}
	if policy.Labels != "{}" {
		t.Fatalf("labels = %q, want {}", policy.Labels)
	}
}

func TestNormalizePolicyRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*bo.AlertEscalationPolicy)
	}{
		{"empty name", func(p *bo.AlertEscalationPolicy) { p.Name = "  " }},
		{"long name", func(p *bo.AlertEscalationPolicy) { p.Name = strings.Repeat("告", 65) }},
		{"long description", func(p *bo.AlertEscalationPolicy) { p.Description = strings.Repeat("述", 1025) }},
		{"labels not an object", func(p *bo.AlertEscalationPolicy) { p.Labels = `["env"]` }},
		{"labels bad regex", func(p *bo.AlertEscalationPolicy) { p.Labels = `{"env":"["}` }},
		{"no steps", func(p *bo.AlertEscalationPolicy) { p.Steps = nil }},
		{"too many steps", func(p *bo.AlertEscalationPolicy) {
			p.Steps = make([]bo.AlertEscalationStep, 21)
			for i := range p.Steps {
				p.Steps[i] = bo.AlertEscalationStep{DelayMinutes: i + 1, NotificationGroupUIDs: []string{"group-1"}}
			}
		}},
		{"negative delay", func(p *bo.AlertEscalationPolicy) { p.Steps[0].DelayMinutes = -1 }},
		{"delay too large", func(p *bo.AlertEscalationPolicy) { p.Steps[1].DelayMinutes = maxDelayMinutes + 1 }},
		{"delays not increasing", func(p *bo.AlertEscalationPolicy) { p.Steps[1].DelayMinutes = 5 }},
		{"last delay zero", func(p *bo.AlertEscalationPolicy) {
			p.Steps = []bo.AlertEscalationStep{{DelayMinutes: 0, NotificationGroupUIDs: []string{"group-1"}}}
		}},
		{"empty step groups", func(p *bo.AlertEscalationPolicy) { p.Steps[0].NotificationGroupUIDs = nil }},
		{"too many step groups", func(p *bo.AlertEscalationPolicy) {
			groups := make([]string, 0, maxStepGroups+1)
			for i := 0; i <= maxStepGroups; i++ {
				groups = append(groups, "group-"+strings.Repeat("x", i+1))
			}
			p.Steps[0].NotificationGroupUIDs = groups
		}},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			policy := escalationTestPolicy()
			testCase.mutate(&policy)
			requireInvalid(t, normalizePolicy(&policy))
		})
	}
}

func TestNormalizeLabelsRejectsEmptyKeyOrValue(t *testing.T) {
	for _, raw := range []string{`{"":"x"}`, `{"env":""}`} {
		if _, err := normalizeLabels(raw); err == nil {
			t.Fatalf("normalizeLabels(%q) should fail", raw)
		}
	}
}
