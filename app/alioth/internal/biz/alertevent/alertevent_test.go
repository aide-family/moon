package alertevent

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/actor"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

func TestValidateHistoryListQuery(t *testing.T) {
	from := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	to := from.Add(time.Hour)
	// 前端一天是 [00:00:00, 23:59:59.999999999]，30 天整正好不超过上限，
	// 第 31 天就必须拒绝。
	thirtyDaysFrom := time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC)
	thirtyDaysTo := time.Date(2026, 10, 30, 23, 59, 59, 999999999, time.UTC)
	thirtyOneDaysTo := thirtyDaysTo.Add(24 * time.Hour)
	cases := []struct {
		name    string
		query   ListQuery
		wantErr bool
	}{
		{name: "history filters allowed", query: ListQuery{View: "HISTORY", StartsAtFrom: &from, StartsAtTo: &to, EventStatus: "RECOVERED", Suppression: "SUPPRESSED", RuleName: "API", DatasourceName: "prod"}},
		{name: "history-only filter rejected in realtime", query: ListQuery{View: "REALTIME", RuleName: "API"}, wantErr: true},
		{name: "invalid time range", query: ListQuery{View: "HISTORY", StartsAtFrom: &to, StartsAtTo: &from}, wantErr: true},
		{name: "30 days allowed", query: ListQuery{View: "HISTORY", StartsAtFrom: &thirtyDaysFrom, StartsAtTo: &thirtyDaysTo}},
		{name: "31 days rejected", query: ListQuery{View: "HISTORY", StartsAtFrom: &thirtyDaysFrom, StartsAtTo: &thirtyOneDaysTo}, wantErr: true},
		{name: "unbounded allowed", query: ListQuery{View: "HISTORY", StartsAtFrom: &from}},
		{name: "invalid status", query: ListQuery{View: "HISTORY", EventStatus: "UNKNOWN"}, wantErr: true},
		{name: "invalid suppression filter", query: ListQuery{View: "HISTORY", Suppression: "MAYBE"}, wantErr: true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateHistoryListQuery(tc.query)
			if (err != nil) != tc.wantErr {
				t.Fatalf("validateHistoryListQuery() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestChartEnd(t *testing.T) {
	now := time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC)
	startsAt := now.Add(-48 * time.Hour)
	recoveredAt := now.Add(-24 * time.Hour)
	future := now.Add(time.Hour)

	cases := []struct {
		name  string
		event bo.AlertEvent
		want  time.Time
	}{
		{
			name:  "recovered event ends at recovery time",
			event: bo.AlertEvent{StartsAt: startsAt, EndsAt: &recoveredAt},
			want:  recoveredAt,
		},
		{
			name:  "firing event ends at now",
			event: bo.AlertEvent{StartsAt: startsAt},
			want:  now,
		},
		{
			name:  "zero ends_at ends at now",
			event: bo.AlertEvent{StartsAt: startsAt, EndsAt: &time.Time{}},
			want:  now,
		},
		{
			name:  "future ends_at falls back to now",
			event: bo.AlertEvent{StartsAt: startsAt, EndsAt: &future},
			want:  now,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := chartEnd(tc.event, now); !got.Equal(tc.want) {
				t.Fatalf("chartEnd() = %s, want %s", got, tc.want)
			}
		})
	}
}

func TestChartWindow(t *testing.T) {
	now := time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)
	start, end, step := chartWindow(now.Add(-time.Minute), now)
	if start != now.Add(-time.Minute).Unix() || end != now.Unix() || step != 15 {
		t.Fatalf("short window = %d %d %d", start, end, step)
	}

	start, end, step = chartWindow(now.Add(-24*time.Hour), now)
	if end-start > 300*step {
		t.Fatalf("too many points: span=%d step=%d", end-start, step)
	}
	if step%15 != 0 {
		t.Fatalf("step = %d, want multiple of 15", step)
	}
	if start >= end {
		t.Fatalf("window = %d..%d", start, end)
	}

	start, end, step = chartWindow(now.Add(time.Hour), now)
	if end != start+1 || step != 15 {
		t.Fatalf("future start = %d %d %d", start, end, step)
	}
}

// fixNamespaceRepo 让测试操作者成为启用租户的负责人，避免用例依赖角色数据。
type fixNamespaceRepo struct{ repository.NamespaceRepo }

func (fixNamespaceRepo) GetNamespace(_ context.Context, uid string) (*bo.Namespace, error) {
	return &bo.Namespace{UID: uid, Name: uid, OwnerUID: "fix-owner", Status: "ACTIVE"}, nil
}

// fixFeatureRepo 返回完整的功能与能力集合，让介入入口的权限检查通过。
type fixFeatureRepo struct{ repository.FeatureRepo }

func (fixFeatureRepo) EffectiveModules(context.Context, string) ([]string, error) {
	return perm.FeatureCodes(), nil
}

func (fixFeatureRepo) EffectiveCapabilities(context.Context, string) ([]string, error) {
	items := perm.Capabilities()
	codes := make([]string, 0, len(items))
	for _, item := range items {
		codes = append(codes, item.Code)
	}
	return codes, nil
}

// fixAlertEventRepo 用 existing 模拟租户内存在的事件，并记录 Intervene 是否被调用。
type fixAlertEventRepo struct {
	repository.AlertEventRepo
	existing       map[string]bool
	interveneCalls int
	intervenedUIDs []string
}

func (r *fixAlertEventRepo) GetEvent(_ context.Context, _, uid string) (*bo.AlertEvent, error) {
	if !r.existing[uid] {
		return nil, merr.NotFound("not found")
	}
	return &bo.AlertEvent{UID: uid}, nil
}

func (r *fixAlertEventRepo) Intervene(_ context.Context, _ string, uids []string, _ string) error {
	r.interveneCalls++
	r.intervenedUIDs = append([]string(nil), uids...)
	return nil
}

func fixInterveneContext() context.Context {
	return actor.WithUser(actor.WithNamespace(context.Background(), "fix-ns"), "fix-owner")
}

func fixInterveneCoordinator(events *fixAlertEventRepo) *Coordinator {
	checker := access.Checker{Namespaces: fixNamespaceRepo{}, Features: fixFeatureRepo{}, Log: slog.Default()}
	return NewCoordinator(nil, events, nil, nil, nil, nil, nil, nil, checker, slog.Default())
}

// TestInterveneUnknownUIDReturnsNotFound 复现 M2/M3：不存在的 uid 不能静默成功。
func TestInterveneUnknownUIDReturnsNotFound(t *testing.T) {
	events := &fixAlertEventRepo{existing: map[string]bool{}}
	err := fixInterveneCoordinator(events).Intervene(fixInterveneContext(), []string{"missing-event"})
	if !merr.IsNotFound(err) {
		t.Fatalf("Intervene error = %v, want not found", err)
	}
	if events.interveneCalls != 0 {
		t.Fatalf("repo.Intervene was called %d times", events.interveneCalls)
	}
}

// TestInterveneKnownUIDsUpdates 校验全部存在时仍然按原逻辑更新一次。
func TestInterveneKnownUIDsUpdates(t *testing.T) {
	events := &fixAlertEventRepo{existing: map[string]bool{"event-1": true, "event-2": true}}
	if err := fixInterveneCoordinator(events).Intervene(fixInterveneContext(), []string{"event-1", "event-2"}); err != nil {
		t.Fatalf("Intervene error = %v", err)
	}
	if events.interveneCalls != 1 || len(events.intervenedUIDs) != 2 || events.intervenedUIDs[0] != "event-1" || events.intervenedUIDs[1] != "event-2" {
		t.Fatalf("calls = %d, uids = %#v", events.interveneCalls, events.intervenedUIDs)
	}
}

// TestInterveneMixedUnknownUIDDoesNotUpdate 校验先全量校验再更新：混合未知 uid 时不能部分更新。
func TestInterveneMixedUnknownUIDDoesNotUpdate(t *testing.T) {
	events := &fixAlertEventRepo{existing: map[string]bool{"event-1": true}}
	err := fixInterveneCoordinator(events).Intervene(fixInterveneContext(), []string{"event-1", "missing-event"})
	if !merr.IsNotFound(err) {
		t.Fatalf("Intervene error = %v, want not found", err)
	}
	if events.interveneCalls != 0 {
		t.Fatalf("repo.Intervene was called %d times", events.interveneCalls)
	}
}
