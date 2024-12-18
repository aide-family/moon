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
}
