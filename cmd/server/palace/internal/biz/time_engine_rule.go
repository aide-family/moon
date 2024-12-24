package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// NewTimeEngineRuleBiz .
func NewTimeEngineRuleBiz(timeEngineRuleRepository repository.TimeEngineRule) *TimeEngineRuleBiz {
	return &TimeEngineRuleBiz{
		timeEngineRuleRepository: timeEngineRuleRepository,
	}
}

// TimeEngineRuleBiz .
type TimeEngineRuleBiz struct {
	timeEngineRuleRepository repository.TimeEngineRule
}

// CreateTimeEngineRule 创建时间引擎规则
func (b *TimeEngineRuleBiz) CreateTimeEngineRule(ctx context.Context, req *bo.CreateTimeEngineRuleRequest) error {
	timeEngineRule := req.Do(ctx)
	return b.timeEngineRuleRepository.CreateTimeEngineRule(ctx, timeEngineRule)
}

// UpdateTimeEngineRule 更新时间引擎规则
func (b *TimeEngineRuleBiz) UpdateTimeEngineRule(ctx context.Context, req *bo.UpdateTimeEngineRuleRequest) error {
	timeEngineRule := req.Do(ctx)
	return b.timeEngineRuleRepository.UpdateTimeEngineRule(ctx, timeEngineRule)
}

// DeleteTimeEngineRule 删除时间引擎规则
func (b *TimeEngineRuleBiz) DeleteTimeEngineRule(ctx context.Context, req *bo.DeleteTimeEngineRuleRequest) error {
	return b.timeEngineRuleRepository.DeleteTimeEngineRule(ctx, req.ID)
}

// GetTimeEngineRule 获取时间引擎规则
func (b *TimeEngineRuleBiz) GetTimeEngineRule(ctx context.Context, req *bo.GetTimeEngineRuleRequest) (*bizmodel.TimeEngineRule, error) {
	return b.timeEngineRuleRepository.GetTimeEngineRule(ctx, req.ID)
}

// ListTimeEngineRule 获取时间引擎规则列表
func (b *TimeEngineRuleBiz) ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) ([]*bizmodel.TimeEngineRule, error) {
	return b.timeEngineRuleRepository.ListTimeEngineRule(ctx, req)
}

// BatchUpdateTimeEngineRuleStatus 批量更新时间引擎规则状态
func (b *TimeEngineRuleBiz) BatchUpdateTimeEngineRuleStatus(ctx context.Context, req *bo.BatchUpdateTimeEngineRuleStatusRequest) error {
	return b.timeEngineRuleRepository.BatchUpdateTimeEngineRuleStatus(ctx, req)
}

// GetTimeEngineRuleByID 获取时间引擎规则
func (b *TimeEngineRuleBiz) GetTimeEngineRuleByID(ctx context.Context, id uint32) (*bizmodel.TimeEngineRule, error) {
	return b.timeEngineRuleRepository.GetTimeEngineRule(ctx, id)
}

// CreateTimeEngine 创建时间引擎
func (b *TimeEngineRuleBiz) CreateTimeEngine(ctx context.Context, req *bo.CreateTimeEngineRequest) error {
	timeEngine := req.Do(ctx)
	return b.timeEngineRuleRepository.CreateTimeEngine(ctx, timeEngine)
}

// UpdateTimeEngine 更新时间引擎
func (b *TimeEngineRuleBiz) UpdateTimeEngine(ctx context.Context, req *bo.UpdateTimeEngineRequest) error {
	timeEngine := req.Do(ctx)
	return b.timeEngineRuleRepository.UpdateTimeEngine(ctx, timeEngine)
}

// DeleteTimeEngine 删除时间引擎
func (b *TimeEngineRuleBiz) DeleteTimeEngine(ctx context.Context, req *bo.DeleteTimeEngineRequest) error {
	return b.timeEngineRuleRepository.DeleteTimeEngine(ctx, req.ID)
}

// GetTimeEngine 获取时间引擎
func (b *TimeEngineRuleBiz) GetTimeEngine(ctx context.Context, req *bo.GetTimeEngineRequest) (*bizmodel.TimeEngine, error) {
	return b.timeEngineRuleRepository.GetTimeEngine(ctx, req.ID)
}

// ListTimeEngine 获取时间引擎列表
func (b *TimeEngineRuleBiz) ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) ([]*bizmodel.TimeEngine, error) {
	return b.timeEngineRuleRepository.ListTimeEngine(ctx, req)
}

// BatchUpdateTimeEngineStatus 批量更新时间引擎状态
func (b *TimeEngineRuleBiz) BatchUpdateTimeEngineStatus(ctx context.Context, req *bo.BatchUpdateTimeEngineStatusRequest) error {
	return b.timeEngineRuleRepository.BatchUpdateTimeEngineStatus(ctx, req)
}
