package cluster

import (
	"context"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/aide-family/moon/app/alioth/internal/access"
	"github.com/aide-family/moon/app/alioth/internal/data/repository"
	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/id"
	"github.com/aide-family/moon/app/alioth/pkg/perm"
	"github.com/aide-family/moon/pkg/merr"
)

const (
	// podExecTicketTTL 是票据的有效期：浏览器拿到票据后要立刻建立 WebSocket，
	// 过期后必须重新申请，避免票据长期可重放。
	podExecTicketTTL = 60 * time.Second
	// maxPodExecTargetLength 与 proto 校验一致，避免绕过 HTTP 的调用写入超长名称。
	maxPodExecTargetLength = 253
	// defaultPodExecCols 和 defaultPodExecRows 是未指定窗口大小时的默认值。
	defaultPodExecCols = int32(80)
	defaultPodExecRows = int32(24)
	// maxPodExecTerminalSize 与 proto 校验一致，避免绕过 HTTP 的调用要求超大窗口。
	maxPodExecTerminalSize = int32(1000)
)

// podExecTicket 是一次 Pod 终端票据绑定的目标。
// 窗口大小也一起绑定，浏览器建连后不需要再传 cols/rows。
type podExecTicket struct {
	clusterUID string
	namespace  string
	name       string
	container  string
	cols, rows int32
	expires    time.Time
}

// Terminal 是 Pod 终端的用例层：先签发一次性票据，再用票据换 WebSocket 连接。
// 票据同时绑定集群、命名空间、Pod、容器和窗口大小；重复使用、过期或目标不匹配
// 一律按未认证处理，不暴露票据是否存在。
type Terminal struct {
	clusters  repository.ClusterRepo
	resources repository.KubernetesRepo
	access    access.Checker

	mu      sync.Mutex
	tickets map[string]podExecTicket
	shells  map[*bo.KubePodShell]struct{}
}

func NewTerminal(clusters repository.ClusterRepo, resources repository.KubernetesRepo, checker access.Checker) *Terminal {
	return &Terminal{
		clusters:  clusters,
		resources: resources,
		access:    checker,
		tickets:   make(map[string]podExecTicket),
		shells:    make(map[*bo.KubePodShell]struct{}),
	}
}

// IssueTicket 签发 60 秒内有效的一次性票据。权限和租户归属都在这里校验，
// 消费票据时不再重复校验：浏览器 WebSocket 无法带 Authorization 头，票据就是建连凭证。
func (terminal *Terminal) IssueTicket(ctx context.Context, clusterUID, namespace, name, container string, cols, rows int32) (string, error) {
	ns, err := terminal.access.Allow(ctx, perm.K8sPodExec)
	if err != nil {
		return "", err
	}
	clusterUID, namespace, name, container, err = normalizePodExecTarget(clusterUID, namespace, name, container)
	if err != nil {
		return "", err
	}
	// 跨租户按不存在处理：集群不属于当前租户时签不出任何票据。
	if _, err := terminal.clusters.GetCluster(ctx, ns, clusterUID); err != nil {
		return "", terminal.access.Wrap(err)
	}
	return terminal.issueTicket(clusterUID, namespace, name, container, cols, rows), nil
}

// issueTicket 只做内存写入，不校验权限。IssueTicket 和单测共用同一份绑定与默认值逻辑。
func (terminal *Terminal) issueTicket(clusterUID, namespace, name, container string, cols, rows int32) string {
	cols, rows = normalizeTerminalSize(cols, rows)
	ticket := id.New()
	now := time.Now()
	terminal.mu.Lock()
	defer terminal.mu.Unlock()
	// 顺手清掉过期票据，避免长期运行的进程里票据表只增不减。
	for key, item := range terminal.tickets {
		if now.After(item.expires) {
			delete(terminal.tickets, key)
		}
	}
	terminal.tickets[ticket] = podExecTicket{
		clusterUID: clusterUID,
		namespace:  namespace,
		name:       name,
		container:  container,
		cols:       cols,
		rows:       rows,
		expires:    now.Add(podExecTicketTTL),
	}
	return ticket
}

// OpenShell 消费票据并打开终端。票据只能消费一次；容器和窗口大小可以留空，
// 此时沿用签发时绑定的值。建连失败不补发票据，页面需要重新申请。
func (terminal *Terminal) OpenShell(ctx context.Context, clusterUID, namespace, name, container, ticket string, cols, rows int32) (*bo.KubePodShell, error) {
	target, err := terminal.takeTicket(clusterUID, namespace, name, container, ticket)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(container) == "" {
		container = target.container
	}
	if cols <= 0 {
		cols = target.cols
	}
	if rows <= 0 {
		rows = target.rows
	}
	cols, rows = normalizeTerminalSize(cols, rows)
	shell, err := terminal.resources.ExecPod(ctx, target.clusterUID, target.namespace, target.name, container, cols, rows)
	if err != nil {
		return nil, terminal.access.Wrap(err)
	}
	terminal.track(shell)
	return shell, nil
}

// takeTicket 原子地取走票据并校验绑定。取走即删除，重复使用会走到 not found。
func (terminal *Terminal) takeTicket(clusterUID, namespace, name, container, ticket string) (podExecTicket, error) {
	clusterUID, namespace = strings.TrimSpace(clusterUID), strings.TrimSpace(namespace)
	name, container = strings.TrimSpace(name), strings.TrimSpace(container)
	ticket = strings.TrimSpace(ticket)
	terminal.mu.Lock()
	item, ok := terminal.tickets[ticket]
	if ok {
		delete(terminal.tickets, ticket)
	}
	terminal.mu.Unlock()
	if !ok || time.Now().After(item.expires) {
		return podExecTicket{}, merr.Unauthenticated("invalid ticket")
	}
	if item.clusterUID != clusterUID || item.namespace != namespace || item.name != name {
		return podExecTicket{}, merr.Unauthenticated("invalid ticket")
	}
	// 建连路径可以不带 container：容器和窗口大小由票据带入；显式传了就必须一致。
	if container != "" && item.container != container {
		return podExecTicket{}, merr.Unauthenticated("invalid ticket")
	}
	return item, nil
}

// track 记录活动终端，供进程退出时统一关闭。包装后的 Stop 会先摘除记录再关闭会话，
// 避免 CloseAll 与 HTTP 处理器的 defer 重复关闭同一会话。
func (terminal *Terminal) track(shell *bo.KubePodShell) {
	stop := shell.Stop
	if stop == nil {
		stop = func() {}
	}
	terminal.mu.Lock()
	terminal.shells[shell] = struct{}{}
	terminal.mu.Unlock()
	shell.Stop = func() {
		terminal.mu.Lock()
		delete(terminal.shells, shell)
		terminal.mu.Unlock()
		stop()
	}
}

// CloseAll 关闭全部活动终端，由 kratos.BeforeStop 在进程退出时调用。
func (terminal *Terminal) CloseAll() {
	terminal.mu.Lock()
	shells := make([]*bo.KubePodShell, 0, len(terminal.shells))
	for shell := range terminal.shells {
		shells = append(shells, shell)
	}
	terminal.mu.Unlock()
	for _, shell := range shells {
		shell.Stop()
	}
}

// normalizePodExecTarget 清理并校验终端目标。集群、命名空间、Pod、容器都必须非空。
func normalizePodExecTarget(clusterUID, namespace, name, container string) (string, string, string, string, error) {
	clusterUID = strings.TrimSpace(clusterUID)
	namespace = strings.TrimSpace(namespace)
	name = strings.TrimSpace(name)
	container = strings.TrimSpace(container)
	if clusterUID == "" {
		return "", "", "", "", merr.Invalid("cluster is required")
	}
	if namespace == "" {
		return "", "", "", "", merr.Invalid("namespace is required")
	}
	if name == "" {
		return "", "", "", "", merr.Invalid("pod name is required")
	}
	if container == "" {
		return "", "", "", "", merr.Invalid("container is required")
	}
	for _, value := range []string{namespace, name, container} {
		if utf8.RuneCountInString(value) > maxPodExecTargetLength {
			return "", "", "", "", merr.Invalid("invalid request")
		}
	}
	return clusterUID, namespace, name, container, nil
}

// normalizeTerminalSize 补齐并收敛窗口大小。0 表示按默认 80x24 处理。
func normalizeTerminalSize(cols, rows int32) (int32, int32) {
	if cols <= 0 {
		cols = defaultPodExecCols
	}
	if rows <= 0 {
		rows = defaultPodExecRows
	}
	if cols > maxPodExecTerminalSize {
		cols = maxPodExecTerminalSize
	}
	if rows > maxPodExecTerminalSize {
		rows = maxPodExecTerminalSize
	}
	return cols, rows
}
