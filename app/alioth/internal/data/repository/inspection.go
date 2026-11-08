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
}
