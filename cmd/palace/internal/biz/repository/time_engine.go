package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
)

// TimeEngine 时间引擎仓储接口
type TimeEngine interface {
	// CreateTimeEngine 创建时间引擎
	CreateTimeEngine(ctx context.Context, req *bo.SaveTimeEngineRequest) error
	// UpdateTimeEngine 更新时间引擎
	UpdateTimeEngine(ctx context.Context, timeEngineId uint32, req *bo.SaveTimeEngineRequest) error
	// UpdateTimeEngineStatus 更新时间引擎状态
	UpdateTimeEngineStatus(ctx context.Context, req *bo.UpdateTimeEngineStatusRequest) error
	// DeleteTimeEngine 删除时间引擎
	DeleteTimeEngine(ctx context.Context, req *bo.DeleteTimeEngineRequest) error
	// GetTimeEngine 获取时间引擎详情
	GetTimeEngine(ctx context.Context, req *bo.GetTimeEngineRequest) (do.TimeEngine, error)
	// ListTimeEngine 获取时间引擎列表
	ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) (*bo.ListTimeEngineReply, error)
}

// TimeEngineRule 时间引擎规则仓储接口
type TimeEngineRule interface {
	// CreateTimeEngineRule 创建时间引擎规则
	CreateTimeEngineRule(ctx context.Context, req *bo.SaveTimeEngineRuleRequest) error
	// UpdateTimeEngineRule 更新时间引擎规则
	UpdateTimeEngineRule(ctx context.Context, timeEngineRuleId uint32, req *bo.SaveTimeEngineRuleRequest) error
	// UpdateTimeEngineRuleStatus 更新时间引擎规则状态
	UpdateTimeEngineRuleStatus(ctx context.Context, req *bo.UpdateTimeEngineRuleStatusRequest) error
	// DeleteTimeEngineRule 删除时间引擎规则
	DeleteTimeEngineRule(ctx context.Context, req *bo.DeleteTimeEngineRuleRequest) error
	// GetTimeEngineRule 获取时间引擎规则详情
	GetTimeEngineRule(ctx context.Context, req *bo.GetTimeEngineRuleRequest) (do.TimeEngineRule, error)
	// ListTimeEngineRule 获取时间引擎规则列表
	ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) (*bo.ListTimeEngineRuleReply, error)
	// Find 获取时间引擎规则列表
	// 如果ruleIds为空，则返回所有时间引擎规则
	Find(ctx context.Context, ruleIds ...uint32) ([]do.TimeEngineRule, error)
}
