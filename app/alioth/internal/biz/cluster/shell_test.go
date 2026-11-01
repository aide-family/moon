package cluster

import (
	"context"
	"testing"
	"time"

	kerrors "github.com/go-kratos/kratos/v3/errors"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/pkg/merr"
)

// fakeExecResources 只实现 ExecPod；其余仓储方法用 nil 嵌入接口，误调用会 panic。
type fakeExecResources struct {
	repository.KubernetesRepo
	exec func(ctx context.Context, clusterUID, namespace, name, container string, cols, rows int32) (*bo.KubePodShell, error)
}

func (fake *fakeExecResources) ExecPod(ctx context.Context, clusterUID, namespace, name, container string, cols, rows int32) (*bo.KubePodShell, error) {
	return fake.exec(ctx, clusterUID, namespace, name, container, cols, rows)
}

func newTestTerminal(exec func(ctx context.Context, clusterUID, namespace, name, container string, cols, rows int32) (*bo.KubePodShell, error)) *Terminal {
	return NewTerminal(nil, &fakeExecResources{exec: exec}, access.Checker{})
}

func requireUnauthenticated(t *testing.T, err error) {
	t.Helper()
	if err == nil || kerrors.Reason(err) != merr.ReasonUnauthenticated {
		t.Fatalf("err = %v, want unauthenticated", err)
	}
}

// TestIssueTicketRequiresPermission 确认签发入口先过权限检查。
// 这里三个依赖都不注入：漏掉 access.Allow 会直接因 nil 仓储 panic，而不是返回未认证。
func TestIssueTicketRequiresPermission(t *testing.T) {
	terminal := NewTerminal(nil, nil, access.Checker{})
	_, err := terminal.IssueTicket(context.Background(), "cluster-1", "default", "web-0", "app", 80, 24)
	if err == nil {
		t.Fatal("IssueTicket without tenant context should fail")
	}
	if kerrors.Reason(err) != merr.ReasonClient {
		t.Fatalf("reason = %q, want client error from missing tenant", kerrors.Reason(err))
	}
	if len(terminal.tickets) != 0 {
		t.Fatalf("tickets = %d, want 0", len(terminal.tickets))
	}
}

// TestTerminalTicketOneTimeUse 确认同一张票据只能开一次终端。
func TestTerminalTicketOneTimeUse(t *testing.T) {
	var opened int
	terminal := newTestTerminal(func(context.Context, string, string, string, string, int32, int32) (*bo.KubePodShell, error) {
		opened++
		return &bo.KubePodShell{Container: "app"}, nil
	})
	ticket := terminal.issueTicket("cluster-1", "default", "web-0", "app", 80, 24)
	if _, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", ticket, 0, 0); err != nil {
		t.Fatalf("first OpenShell: %v", err)
	}
	_, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", ticket, 0, 0)
	requireUnauthenticated(t, err)
	if opened != 1 {
		t.Fatalf("ExecPod calls = %d, want 1", opened)
	}
}

// TestTerminalTicketExpired 确认过期票据被拒绝，即使目标完全匹配。
func TestTerminalTicketExpired(t *testing.T) {
	terminal := newTestTerminal(func(context.Context, string, string, string, string, int32, int32) (*bo.KubePodShell, error) {
		t.Fatal("expired ticket must not open a shell")
		return nil, nil
	})
	ticket := terminal.issueTicket("cluster-1", "default", "web-0", "app", 80, 24)
	terminal.mu.Lock()
	item := terminal.tickets[ticket]
	item.expires = time.Now().Add(-time.Second)
	terminal.tickets[ticket] = item
	terminal.mu.Unlock()
	_, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", ticket, 0, 0)
	requireUnauthenticated(t, err)
}

// TestTerminalTicketTargetMismatch 确认票据绑定集群、命名空间、Pod 和容器。
func TestTerminalTicketTargetMismatch(t *testing.T) {
	terminal := newTestTerminal(func(context.Context, string, string, string, string, int32, int32) (*bo.KubePodShell, error) {
		t.Fatal("mismatched ticket must not open a shell")
		return nil, nil
	})
	cases := []struct {
		name      string
		cluster   string
		namespace string
		pod       string
		container string
	}{
		{name: "other cluster", cluster: "cluster-2", namespace: "default", pod: "web-0"},
		{name: "other namespace", cluster: "cluster-1", namespace: "kube-system", pod: "web-0"},
		{name: "other pod", cluster: "cluster-1", namespace: "default", pod: "web-1"},
		{name: "other container", cluster: "cluster-1", namespace: "default", pod: "web-0", container: "sidecar"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ticket := terminal.issueTicket("cluster-1", "default", "web-0", "app", 80, 24)
			_, err := terminal.OpenShell(context.Background(), tc.cluster, tc.namespace, tc.pod, tc.container, ticket, 0, 0)
			requireUnauthenticated(t, err)
		})
	}
	if _, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", "", 0, 0); err == nil {
		// 空票据也必须按未认证处理，不能命中任何目标。
		t.Fatal("empty ticket should fail")
	} else {
		requireUnauthenticated(t, err)
	}
}

// TestTerminalTicketDefaults 确认窗口大小的默认值和优先级：请求参数 > 票据绑定 > 80x24。
func TestTerminalTicketDefaults(t *testing.T) {
	type call struct {
		container  string
		cols, rows int32
	}
	calls := make([]call, 0, 3)
	terminal := newTestTerminal(func(_ context.Context, _, _, _, container string, cols, rows int32) (*bo.KubePodShell, error) {
		calls = append(calls, call{container: container, cols: cols, rows: rows})
		return &bo.KubePodShell{Container: container}, nil
	})
	// 票据未指定窗口大小时存默认 80x24，建连也不再传，落到默认值。
	ticket := terminal.issueTicket("cluster-1", "default", "web-0", "app", 0, 0)
	if _, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", ticket, 0, 0); err != nil {
		t.Fatalf("default OpenShell: %v", err)
	}
	// 票据绑定的窗口大小在未显式传参时生效。
	ticket = terminal.issueTicket("cluster-1", "default", "web-0", "app", 120, 40)
	if _, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", ticket, 0, 0); err != nil {
		t.Fatalf("ticket size OpenShell: %v", err)
	}
	// 显式传参优先于票据绑定。
	ticket = terminal.issueTicket("cluster-1", "default", "web-0", "app", 120, 40)
	if _, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", ticket, 100, 50); err != nil {
		t.Fatalf("explicit size OpenShell: %v", err)
	}
	want := []call{
		{container: "app", cols: 80, rows: 24},
		{container: "app", cols: 120, rows: 40},
		{container: "app", cols: 100, rows: 50},
	}
	if len(calls) != len(want) {
		t.Fatalf("calls = %+v, want %+v", calls, want)
	}
	for index := range want {
		if calls[index] != want[index] {
			t.Fatalf("call %d = %+v, want %+v", index, calls[index], want[index])
		}
	}
}

// TestTerminalCloseAllStopsShells 确认进程退出时统一关闭活动终端，且 Stop 只执行一次。
func TestTerminalCloseAllStopsShells(t *testing.T) {
	stops := 0
	terminal := newTestTerminal(func(_ context.Context, _, _, _, container string, _, _ int32) (*bo.KubePodShell, error) {
		return &bo.KubePodShell{Container: container, Stop: func() { stops++ }}, nil
	})
	first := terminal.issueTicket("cluster-1", "default", "web-0", "app", 80, 24)
	if _, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-0", "", first, 0, 0); err != nil {
		t.Fatalf("first OpenShell: %v", err)
	}
	second := terminal.issueTicket("cluster-1", "default", "web-1", "app", 80, 24)
	if _, err := terminal.OpenShell(context.Background(), "cluster-1", "default", "web-1", "", second, 0, 0); err != nil {
		t.Fatalf("second OpenShell: %v", err)
	}
	if len(terminal.shells) != 2 {
		t.Fatalf("tracked shells = %d, want 2", len(terminal.shells))
	}
	terminal.CloseAll()
	if stops != 2 {
		t.Fatalf("stops = %d, want 2", stops)
	}
	if len(terminal.shells) != 0 {
		t.Fatalf("tracked shells after CloseAll = %d, want 0", len(terminal.shells))
	}
	// 再次关闭不能重复 Stop 数据层会话。
	terminal.CloseAll()
	if stops != 2 {
		t.Fatalf("stops after second CloseAll = %d, want 2", stops)
	}
}
