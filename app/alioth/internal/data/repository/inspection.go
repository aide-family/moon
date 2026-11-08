package repository

import (
	"context"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// InspectionRepo 管理巡检计划、检查项、报告模板、执行记录与报告快照。
type InspectionRepo interface {
	CreateInspectionPlan(ctx context.Context, in bo.InspectionPlan) (*bo.InspectionPlan, error)
	UpdateInspectionPlan(ctx context.Context, in bo.InspectionPlan) (*bo.InspectionPlan, error)
	UpdateInspectionPlanStatus(ctx context.Context, namespaceUID, uid, status string, nextRunAt *time.Time) (*bo.InspectionPlan, error)
	DeleteInspectionPlan(ctx context.Context, namespaceUID, uid string) error
	GetInspectionPlan(ctx context.Context, namespaceUID, uid string) (*bo.InspectionPlan, error)
	ListInspectionPlans(ctx context.Context, namespaceUID string, filter bo.InspectionPlanListFilter, page, pageSize int) ([]bo.InspectionPlan, int, error)

	CreateInspectionTemplate(ctx context.Context, in bo.InspectionTemplate) (*bo.InspectionTemplate, error)
	UpdateInspectionTemplate(ctx context.Context, in bo.InspectionTemplate) (*bo.InspectionTemplate, error)
	DeleteInspectionTemplate(ctx context.Context, namespaceUID, uid string) error
	GetInspectionTemplate(ctx context.Context, namespaceUID, uid string) (*bo.InspectionTemplate, error)
	ListInspectionTemplates(ctx context.Context, namespaceUID, name string, page, pageSize int) ([]bo.InspectionTemplate, int, error)

	// DueInspectionPlans 返回已到执行时间且启用中的计划，含检查项。
	DueInspectionPlans(ctx context.Context, now time.Time, limit int) ([]bo.InspectionPlan, error)
	// AdvanceInspectionPlan 把计划的下次执行从 expectedNext 推进到 next；
	// expectedNext 用于多副本之间的乐观并发，返回 false 表示已被别的副本推进。
	AdvanceInspectionPlan(ctx context.Context, planUID string, expectedNext, next, lastRunAt time.Time) (bool, error)
	// CreateInspectionRun 幂等创建执行记录；同一计划同一计划时刻已存在时返回 nil。
	CreateInspectionRun(ctx context.Context, in bo.InspectionRun) (*bo.InspectionRun, error)
	// HasActiveInspectionRun 判断计划是否已有 PENDING 或 RUNNING 的执行。
	HasActiveInspectionRun(ctx context.Context, namespaceUID, planUID string) (bool, error)
	// ClaimInspectionRun 用 SKIP LOCKED 领取一条待执行记录并写入租约。
	ClaimInspectionRun(ctx context.Context, owner string, now, leaseUntil time.Time) (*bo.InspectionRun, error)
	// StartInspectionRun 写入本次数据窗口。
	StartInspectionRun(ctx context.Context, uid string, windowStart, windowEnd time.Time) error
	// FinishInspectionRun 写入最终状态、计数、摘要与结果快照。
	FinishInspectionRun(ctx context.Context, in bo.InspectionRun) error
	GetInspectionRun(ctx context.Context, namespaceUID, uid string) (*bo.InspectionRun, error)
	ListInspectionRuns(ctx context.Context, namespaceUID string, filter bo.InspectionRunListFilter, page, pageSize int) ([]bo.InspectionRun, int, error)
}
