package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/do"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
)

// TimeEngine 时间引擎用例
type TimeEngine struct {
	timeEngineRepo     repository.TimeEngine
	timeEngineRuleRepo repository.TimeEngineRule
}

// NewTimeEngine 创建时间引擎用例
func NewTimeEngine(timeEngineRepo repository.TimeEngine, timeEngineRuleRepo repository.TimeEngineRule) *TimeEngine {
	return &TimeEngine{
		timeEngineRepo:     timeEngineRepo,
		timeEngineRuleRepo: timeEngineRuleRepo,
	}
}

// SaveTimeEngine 保存时间引擎
func (t *TimeEngine) SaveTimeEngine(ctx context.Context, req *bo.SaveTimeEngineRequest) error {
	rules, err := t.timeEngineRuleRepo.Find(ctx, req.RuleIds...)
	if err != nil {
		return err
	}
	req.WithRules(rules)
	if req.TimeEngineId == 0 {
		if err := req.Validate(); err != nil {
			return err
		}
		return t.timeEngineRepo.CreateTimeEngine(ctx, req)
	}
	timeEngine, err := t.timeEngineRepo.GetTimeEngine(ctx, &bo.GetTimeEngineRequest{
		TimeEngineId: req.TimeEngineId,
	})
	if err != nil {
		return err
	}
	req.WithTimeEngine(timeEngine)
	if err := req.Validate(); err != nil {
		return err
	}
	return t.timeEngineRepo.UpdateTimeEngine(ctx, req.TimeEngineId, req)
}

// UpdateTimeEngineStatus 更新时间引擎状态
func (t *TimeEngine) UpdateTimeEngineStatus(ctx context.Context, req *bo.UpdateTimeEngineStatusRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return t.timeEngineRepo.UpdateTimeEngineStatus(ctx, req)
}

// DeleteTimeEngine 删除时间引擎
func (t *TimeEngine) DeleteTimeEngine(ctx context.Context, req *bo.DeleteTimeEngineRequest) error {
	return t.timeEngineRepo.DeleteTimeEngine(ctx, req)
}

// GetTimeEngine 获取时间引擎详情
func (t *TimeEngine) GetTimeEngine(ctx context.Context, req *bo.GetTimeEngineRequest) (do.TimeEngine, error) {
	return t.timeEngineRepo.GetTimeEngine(ctx, req)
}

// ListTimeEngine 获取时间引擎列表
func (t *TimeEngine) ListTimeEngine(ctx context.Context, req *bo.ListTimeEngineRequest) (*bo.ListTimeEngineReply, error) {
	return t.timeEngineRepo.ListTimeEngine(ctx, req)
}

// SelectTimeEngine 获取时间引擎列表
func (t *TimeEngine) SelectTimeEngine(ctx context.Context, req *bo.SelectTimeEngineRequest) (*bo.SelectTimeEngineReply, error) {
	return t.timeEngineRepo.SelectTimeEngine(ctx, req)
}

// SaveTimeEngineRule 保存时间引擎规则
func (t *TimeEngine) SaveTimeEngineRule(ctx context.Context, req *bo.SaveTimeEngineRuleRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	if req.TimeEngineRuleId == 0 {
		return t.timeEngineRuleRepo.CreateTimeEngineRule(ctx, req)
	}
	return t.timeEngineRuleRepo.UpdateTimeEngineRule(ctx, req.TimeEngineRuleId, req)
}

// UpdateTimeEngineRuleStatus 更新时间引擎规则状态
func (t *TimeEngine) UpdateTimeEngineRuleStatus(ctx context.Context, req *bo.UpdateTimeEngineRuleStatusRequest) error {
	if err := req.Validate(); err != nil {
		return err
	}
	return t.timeEngineRuleRepo.UpdateTimeEngineRuleStatus(ctx, req)
}

// DeleteTimeEngineRule 删除时间引擎规则
func (t *TimeEngine) DeleteTimeEngineRule(ctx context.Context, req *bo.DeleteTimeEngineRuleRequest) error {
	return t.timeEngineRuleRepo.DeleteTimeEngineRule(ctx, req)
}

// GetTimeEngineRule 获取时间引擎规则详情
func (t *TimeEngine) GetTimeEngineRule(ctx context.Context, req *bo.GetTimeEngineRuleRequest) (do.TimeEngineRule, error) {
	return t.timeEngineRuleRepo.GetTimeEngineRule(ctx, req)
}

// ListTimeEngineRule 获取时间引擎规则列表
func (t *TimeEngine) ListTimeEngineRule(ctx context.Context, req *bo.ListTimeEngineRuleRequest) (*bo.ListTimeEngineRuleReply, error) {
	return t.timeEngineRuleRepo.ListTimeEngineRule(ctx, req)
}

// SelectTimeEngineRule 获取时间引擎规则列表
func (t *TimeEngine) SelectTimeEngineRule(ctx context.Context, req *bo.SelectTimeEngineRuleRequest) (*bo.SelectTimeEngineRuleReply, error) {
	return t.timeEngineRuleRepo.SelectTimeEngineRule(ctx, req)
}
