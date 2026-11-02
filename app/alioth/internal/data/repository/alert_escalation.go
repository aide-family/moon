package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// AlertEscalationRepo 管理告警升级策略、策略步骤、事件升级记录，以及调度所需的
// 补齐、取消与发送。调度方法不区分租户，由调度器在锁内调用。
type AlertEscalationRepo interface {
	CreatePolicy(ctx context.Context, in bo.AlertEscalationPolicy) (*bo.AlertEscalationPolicy, error)
	UpdatePolicy(ctx context.Context, in bo.AlertEscalationPolicy) (*bo.AlertEscalationPolicy, error)
	UpdatePolicyStatus(ctx context.Context, namespaceUID, uid, status string) (*bo.AlertEscalationPolicy, error)
	DeletePolicy(ctx context.Context, namespaceUID, uid string) error
	GetPolicy(ctx context.Context, namespaceUID, uid string) (*bo.AlertEscalationPolicy, error)
	ListPolicies(ctx context.Context, namespaceUID string, filter bo.AlertEscalationPolicyFilter, page, pageSize int) ([]bo.AlertEscalationPolicy, int, error)
	// ListEventEscalations 返回一条事件的全部升级记录，按到期时间升序。
	ListEventEscalations(ctx context.Context, namespaceUID, eventUID string) ([]bo.AlertEventEscalation, error)

	// FillPendingEscalations 给 since 之后开始、仍未认领且未被抑制的 FIRING 事件
	// 补齐命中策略的 PENDING 行；同一步只建一次。
	FillPendingEscalations(ctx context.Context, since, now time.Time) (int, error)
	// CancelStaleEscalations 把事件已认领/恢复/抑制、或策略已停用/删除的 PENDING 行置为 CANCELLED。
	CancelStaleEscalations(ctx context.Context, now time.Time) (int, error)
	// SendDueEscalations 把已到期且仍有效的 PENDING 行生成通知任务：
	// 生成了任务的置 SENT，已有同事件同组同渠道 FIRING 任务的置 SKIPPED。
	SendDueEscalations(ctx context.Context, now time.Time) (int, error)
}
