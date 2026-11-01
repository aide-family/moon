package data

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/internal/data/ent/agent"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sshcommand"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/sshcommandaudit"
	"github.com/aide-family/moon/app/alioth/internal/data/ent/user"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestListAuditsFilters 在真实库上验证审计过滤、时间范围、分页与机器名/命令名补全。
func TestListAuditsFilters(t *testing.T) {
	store := newAgentTestStore(t)
	ctx := context.Background()
	suffix := time.Now().Format("150405.000000")
	namespaceUID := "m25-audit-" + suffix
	otherNamespaceUID := namespaceUID + "-other"
	base := time.Now().Add(-10 * time.Hour).Truncate(time.Second)
	operatorOne := "m25-user-1-" + suffix
	operatorTwo := "m25-user-2-" + suffix

	for _, account := range []struct {
		uid      string
		name     string
		nickname string
		avatar   string
	}{
		{uid: operatorOne, name: "Alice", nickname: "alice", avatar: "alice.png"},
		{uid: operatorTwo, name: "Bob", nickname: "bob", avatar: "bob.png"},
	} {
		if _, err := store.client.User.Create().
			SetUID(account.uid).
			SetEmail(account.uid + "@example.com").
			SetName(account.name).
			SetNickname(account.nickname).
			SetAvatar(account.avatar).
			Save(ctx); err != nil {
			t.Fatalf("create user %s: %v", account.uid, err)
		}
	}
	agents := []struct {
		uid  string
		name string
	}{
		{uid: "m25-audit-agent-1", name: "web-1"},
		{uid: "m25-audit-agent-2", name: "web-2"},
		{uid: "m25-audit-agent-3", name: "legacy-agent"},
	}
	agentUIDs := make([]string, 0, len(agents))
	for _, machine := range agents {
		agentUIDs = append(agentUIDs, machine.uid)
		if _, err := store.client.Agent.Create().
			SetUID(machine.uid).
			SetNamespaceUID(namespaceUID).
			SetName(machine.name).
			Save(ctx); err != nil {
			t.Fatalf("create agent %s: %v", machine.uid, err)
		}
	}
	commands := []struct {
		uid  string
		name string
	}{
		{uid: "m25-audit-command-1", name: "deploy"},
		{uid: "m25-audit-command-2", name: "restart"},
		{uid: "m25-audit-command-3", name: "legacy-command"},
	}
	commandUIDs := make([]string, 0, len(commands))
	for _, command := range commands {
		commandUIDs = append(commandUIDs, command.uid)
		if _, err := store.client.SSHCommand.Create().
			SetUID(command.uid).
			SetNamespaceUID(namespaceUID).
			SetName(command.name).
			SetCommand("echo " + command.name).
			Save(ctx); err != nil {
			t.Fatalf("create command %s: %v", command.uid, err)
		}
	}
	// 机器和命令被删除后审计仍然保留，但名字补全为空。
	now := time.Now()
	if _, err := store.client.Agent.Update().Where(agent.UIDEQ("m25-audit-agent-3")).SetDeletedAt(now).Save(ctx); err != nil {
		t.Fatalf("soft delete agent: %v", err)
	}
	if _, err := store.client.SSHCommand.Update().Where(sshcommand.UIDEQ("m25-audit-command-3")).SetDeletedAt(now).Save(ctx); err != nil {
		t.Fatalf("soft delete command: %v", err)
	}

	audits := []struct {
		uid          string
		namespaceUID string
		commandUID   string
		agentUID     string
		operatorUID  string
		kind         sshcommandaudit.Kind
		status       sshcommandaudit.Status
		command      string
		createdAt    time.Time
	}{
		{uid: "m25-audit-1", namespaceUID: namespaceUID, commandUID: commands[0].uid, agentUID: agents[0].uid, operatorUID: operatorOne, kind: sshcommandaudit.KindEXECUTE, status: sshcommandaudit.StatusDONE, command: "systemctl restart nginx", createdAt: base},
		{uid: "m25-audit-2", namespaceUID: namespaceUID, commandUID: commands[1].uid, agentUID: agents[1].uid, operatorUID: operatorTwo, kind: sshcommandaudit.KindEXECUTE, status: sshcommandaudit.StatusFAILED, command: "echo Hello", createdAt: base.Add(time.Hour)},
		{uid: "m25-audit-3", namespaceUID: namespaceUID, commandUID: commands[0].uid, operatorUID: operatorOne, kind: sshcommandaudit.KindCREATE, status: sshcommandaudit.StatusPENDING, command: "systemctl restart nginx", createdAt: base.Add(2 * time.Hour)},
		{uid: "m25-audit-4", namespaceUID: namespaceUID, commandUID: commands[2].uid, agentUID: agents[2].uid, operatorUID: operatorOne, kind: sshcommandaudit.KindEXECUTE, status: sshcommandaudit.StatusDONE, command: "legacy task", createdAt: base.Add(3 * time.Hour)},
		{uid: "m25-audit-other", namespaceUID: otherNamespaceUID, commandUID: commands[0].uid, agentUID: agents[0].uid, operatorUID: operatorOne, kind: sshcommandaudit.KindEXECUTE, status: sshcommandaudit.StatusDONE, command: "systemctl restart nginx", createdAt: base.Add(4 * time.Hour)},
	}
	auditUIDs := make([]string, 0, len(audits))
	for _, audit := range audits {
		auditUIDs = append(auditUIDs, audit.uid)
		if _, err := store.client.SSHCommandAudit.Create().
			SetUID(audit.uid).
			SetNamespaceUID(audit.namespaceUID).
			SetCommandUID(audit.commandUID).
			SetAgentUID(audit.agentUID).
			SetOperatorUID(audit.operatorUID).
			SetKind(audit.kind).
			SetStatus(audit.status).
			SetCommand(audit.command).
			SetCreatedAt(audit.createdAt).
			Save(ctx); err != nil {
			t.Fatalf("create audit %s: %v", audit.uid, err)
		}
	}
	t.Cleanup(func() {
		cleanupCtx := context.Background()
		_, _ = store.client.SSHCommandAudit.Delete().Where(sshcommandaudit.UIDIn(auditUIDs...)).Exec(cleanupCtx)
		_, _ = store.client.SSHCommand.Delete().Where(sshcommand.NamespaceUIDEQ(namespaceUID)).Exec(cleanupCtx)
		_, _ = store.client.Agent.Delete().Where(agent.UIDIn(agentUIDs...)).Exec(cleanupCtx)
		_, _ = store.client.User.Delete().Where(user.UIDIn(operatorOne, operatorTwo)).Exec(cleanupCtx)
	})
	repo := NewSSHCommandRepo(store, slog.Default())

	items, total, err := repo.ListAudits(ctx, namespaceUID, bo.SSHCommandAuditQuery{}, 1, 20)
	if err != nil {
		t.Fatalf("ListAudits: %v", err)
	}
	if total != 4 || len(items) != 4 {
		t.Fatalf("total = %d, len = %d, want 4", total, len(items))
	}
	if got := auditUIDList(items); got[0] != "m25-audit-4" || got[3] != "m25-audit-1" {
		t.Fatalf("order = %v, want created_at desc", got)
	}
	byUID := make(map[string]bo.SSHCommandAudit, len(items))
	for _, item := range items {
		byUID[item.UID] = item
	}
	if byUID["m25-audit-1"].AgentName != "web-1" || byUID["m25-audit-1"].CommandName != "deploy" {
		t.Fatalf("joined names = %+v", byUID["m25-audit-1"])
	}
	if byUID["m25-audit-1"].OperatorName != "Alice" || byUID["m25-audit-1"].OperatorNickname != "alice" {
		t.Fatalf("operator join = %+v", byUID["m25-audit-1"])
	}
	if byUID["m25-audit-2"].OperatorName != "Bob" {
		t.Fatalf("second operator join = %+v", byUID["m25-audit-2"])
	}
	if byUID["m25-audit-3"].AgentName != "" || byUID["m25-audit-3"].CommandName != "deploy" {
		t.Fatalf("audit without agent = %+v", byUID["m25-audit-3"])
	}
	if byUID["m25-audit-4"].AgentName != "" || byUID["m25-audit-4"].CommandName != "" {
		t.Fatalf("deleted agent/command should leave empty names: %+v", byUID["m25-audit-4"])
	}

	// 页码 2、每页 2 条：total 不变，第二页是较早的两条。
	items, total, err = repo.ListAudits(ctx, namespaceUID, bo.SSHCommandAuditQuery{}, 2, 2)
	if err != nil {
		t.Fatalf("ListAudits page: %v", err)
	}
	if total != 4 || len(items) != 2 || items[0].UID != "m25-audit-2" || items[1].UID != "m25-audit-1" {
		t.Fatalf("page 2 = %+v, total = %d", auditUIDList(items), total)
	}

	filters := []struct {
		name  string
		query bo.SSHCommandAuditQuery
		want  []string
	}{
		{name: "kind", query: bo.SSHCommandAuditQuery{Kind: "EXECUTE"}, want: []string{"m25-audit-4", "m25-audit-2", "m25-audit-1"}},
		{name: "status", query: bo.SSHCommandAuditQuery{Status: "FAILED"}, want: []string{"m25-audit-2"}},
		{name: "operator", query: bo.SSHCommandAuditQuery{OperatorUID: operatorOne}, want: []string{"m25-audit-4", "m25-audit-3", "m25-audit-1"}},
		{name: "command uid", query: bo.SSHCommandAuditQuery{CommandUID: commands[0].uid}, want: []string{"m25-audit-3", "m25-audit-1"}},
		{name: "agent uid", query: bo.SSHCommandAuditQuery{AgentUID: agents[0].uid}, want: []string{"m25-audit-1"}},
		{name: "keyword is case insensitive", query: bo.SSHCommandAuditQuery{Keyword: "HELLO"}, want: []string{"m25-audit-2"}},
		{name: "keyword is trimmed", query: bo.SSHCommandAuditQuery{Keyword: "  nginx "}, want: []string{"m25-audit-3", "m25-audit-1"}},
		{name: "start time inclusive", query: bo.SSHCommandAuditQuery{StartTime: base.Add(90 * time.Minute).Unix()}, want: []string{"m25-audit-4", "m25-audit-3"}},
		{name: "end time exclusive", query: bo.SSHCommandAuditQuery{EndTime: base.Add(90 * time.Minute).Unix()}, want: []string{"m25-audit-2", "m25-audit-1"}},
		{name: "time range", query: bo.SSHCommandAuditQuery{StartTime: base.Add(30 * time.Minute).Unix(), EndTime: base.Add(150 * time.Minute).Unix()}, want: []string{"m25-audit-3", "m25-audit-2"}},
		{name: "execute done", query: bo.SSHCommandAuditQuery{Kind: "EXECUTE", Status: "DONE"}, want: []string{"m25-audit-4", "m25-audit-1"}},
	}
	for _, tt := range filters {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := repo.ListAudits(ctx, namespaceUID, tt.query, 1, 20)
			if err != nil {
				t.Fatalf("ListAudits: %v", err)
			}
			got := auditUIDList(items)
			if total != len(tt.want) || len(got) != len(tt.want) {
				t.Fatalf("total = %d, items = %v, want %v", total, got, tt.want)
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Fatalf("items = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func auditUIDList(items []bo.SSHCommandAudit) []string {
	uids := make([]string, 0, len(items))
	for _, item := range items {
		uids = append(uids, item.UID)
	}
	return uids
}
