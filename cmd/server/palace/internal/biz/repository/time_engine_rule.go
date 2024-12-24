package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// TimeEngineRule .
type TimeEngineRule interface {
	// CreateTimeEngineRule 创建时间引擎规则
	CreateTimeEngineRule(ctx context.Context, req *bizmodel.TimeEngineRule) error

	// UpdateTimeEngineRule 更新时间引擎规则
	UpdateTimeEngineRule(ctx context.Context, req *bizmodel.TimeEngineRule) error

	// DeleteTimeEngineRule 删除时间引擎规则
	DeleteTimeEngineRule(ctx context.Context, id uint32) error

	// GetTimeEngineRule 获取时间引擎规则
	GetTimeEngineRule(ctx context.Context, id uint32) (*bizmodel.TimeEngineRule, error)

	// ListTimeEngineRule 获取时间引擎规则列表
	ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) ([]*bizmodel.TimeEngineRule, error)

	// BatchUpdateTimeEngineRuleStatus 批量更新时间引擎规则状态
	BatchUpdateTimeEngineRuleStatus(ctx context.Context, req *bo.BatchUpdateTimeEngineRuleStatusRequest) error

	// CreateTimeEngine 创建时间引擎
	CreateTimeEngine(ctx context.Context, req *bizmodel.TimeEngine) error

	// UpdateTimeEngine 更新时间引擎
	UpdateTimeEngine(ctx context.Context, req *bizmodel.TimeEngine) error

	// DeleteTimeEngine 删除时间引擎
	DeleteTimeEngine(ctx context.Context, id uint32) error

	// GetTimeEngine 获取时间引擎
	GetTimeEngine(ctx context.Context, id uint32) (*bizmodel.TimeEngine, error)

	// ListTimeEngine 获取时间引擎列表
	ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) ([]*bizmodel.TimeEngine, error)

	// BatchUpdateTimeEngineStatus 批量更新时间引擎状态
	BatchUpdateTimeEngineStatus(ctx context.Context, req *bo.BatchUpdateTimeEngineStatusRequest) error
}
