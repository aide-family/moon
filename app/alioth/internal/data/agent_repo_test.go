package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/conf"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/agent"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestAgentLabelsMatch 覆盖标签匹配的 AND 精确语义与非法 JSON 的跳过行为。
func TestAgentLabelsMatch(t *testing.T) {
	tests := []struct {
		name   string
		raw    string
		labels map[string]string
		want   bool
	}{
		{name: "empty request matches any valid object", raw: `{"env":"prod"}`, labels: nil, want: true},
		{name: "subset matches", raw: `{"env":"prod","role":"web"}`, labels: map[string]string{"env": "prod"}, want: true},
		{name: "all pairs match", raw: `{"env":"prod","role":"web"}`, labels: map[string]string{"env": "prod", "role": "web"}, want: true},
		{name: "missing key", raw: `{"env":"prod"}`, labels: map[string]string{"role": "web"}, want: false},
		{name: "different value", raw: `{"env":"prod"}`, labels: map[string]string{"env": "dev"}, want: false},
		{name: "and requires every pair", raw: `{"env":"prod","role":"web"}`, labels: map[string]string{"env": "prod", "role": "db"}, want: false},
		{name: "empty value is exact", raw: `{"env":""}`, labels: map[string]string{"env": ""}, want: true},
		{name: "invalid json", raw: `not-json`, labels: map[string]string{"env": "prod"}, want: false},
		{name: "empty json text", raw: ``, labels: map[string]string{"env": "prod"}, want: false},
		{name: "json null", raw: `null`, labels: map[string]string{"env": "prod"}, want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := agentLabelsMatch(tt.raw, tt.labels); got != tt.want {
				t.Fatalf("agentLabelsMatch(%q, %v) = %v, want %v", tt.raw, tt.labels, got, tt.want)
			}
		})
	}
}

// TestMatchAgentsByLabels 在真实库上验证租户、ACTIVE 状态和标签 AND 匹配。
func TestMatchAgentsByLabels(t *testing.T) {
	store := newAgentTestStore(t)
	ctx := context.Background()
	namespaceUID := "m25-agent-match-" + time.Now().Format("150405.000000")
	otherNamespaceUID := namespaceUID + "-other"
	rows := []struct {
		uid    string
		ns     string
		status agent.Status
		labels string
	}{
		{uid: "m25-match-1", ns: namespaceUID, status: agent.StatusACTIVE, labels: `{"env":"prod","role":"web"}`},
		{uid: "m25-match-2", ns: namespaceUID, status: agent.StatusACTIVE, labels: `{"env":"prod"}`},
		{uid: "m25-match-3", ns: namespaceUID, status: agent.StatusACTIVE, labels: `{"env":"dev","role":"web"}`},
		{uid: "m25-match-4", ns: namespaceUID, status: agent.StatusREVOKED, labels: `{"env":"prod","role":"web"}`},
		{uid: "m25-match-5", ns: otherNamespaceUID, status: agent.StatusACTIVE, labels: `{"env":"prod","role":"web"}`},
		{uid: "m25-match-6", ns: namespaceUID, status: agent.StatusACTIVE, labels: `not-json`},
	}
	uids := make([]string, 0, len(rows))
	for _, row := range rows {
		uids = append(uids, row.uid)
		if _, err := store.client.Agent.Create().
			SetUID(row.uid).
			SetNamespaceUID(row.ns).
			SetName(row.uid).
			SetStatus(row.status).
			SetLabels(row.labels).
			Save(ctx); err != nil {
			t.Fatalf("create agent %s: %v", row.uid, err)
		}
	}
	t.Cleanup(func() {
		_, _ = store.client.Agent.Delete().Where(agent.UIDIn(uids...)).Exec(context.Background())
	})
	repo := NewAgentRepo(store, slog.Default())

	matched, err := repo.MatchAgents(ctx, namespaceUID, map[string]string{"env": "prod"})
	if err != nil {
		t.Fatalf("MatchAgents: %v", err)
	}
	if got := agentUIDList(matched); len(got) != 2 || got[0] != "m25-match-1" || got[1] != "m25-match-2" {
		t.Fatalf("prod agents = %v", got)
	}
	if matched[0].Labels == "" {
		t.Fatal("matched agent labels are not read back")
	}

	matched, err = repo.MatchAgents(ctx, namespaceUID, map[string]string{"env": "prod", "role": "web"})
	if err != nil {
		t.Fatalf("MatchAgents AND: %v", err)
	}
	if got := agentUIDList(matched); len(got) != 1 || got[0] != "m25-match-1" {
		t.Fatalf("prod web agents = %v", got)
	}

	// 空标签在仓储层表示不过滤，业务层会拒绝这种请求。
	matched, err = repo.MatchAgents(ctx, namespaceUID, map[string]string{})
	if err != nil {
		t.Fatalf("MatchAgents without labels: %v", err)
	}
	if got := agentUIDList(matched); len(got) != 3 {
		t.Fatalf("active agents = %v, want 3 valid rows (invalid labels skipped)", got)
	}
}

// newAgentTestStore 按现有集成测试模式打开 ALIOTH_TEST_DSN 指向的测试库，未设置时跳过。
func newAgentTestStore(t *testing.T) *Dependencies {
	t.Helper()
	source := testDSN(t)
	store, cleanup, err := NewDependencies(
		&conf.Data{Database: &conf.Data_Database{Source: source}},
		&conf.Auth{EncryptKey: "dev-only-encrypt-key"},
		slog.Default(),
	)
	if err != nil {
		t.Skip(err)
	}
	t.Cleanup(cleanup)
	if err := ApplySchema(context.Background(), store.client, store.db); err != nil {
		t.Fatal(err)
	}
	return store
}

func agentUIDList(items []bo.Agent) []string {
	uids := make([]string, 0, len(items))
	for _, item := range items {
		uids = append(uids, item.UID)
	}
	return uids
}
