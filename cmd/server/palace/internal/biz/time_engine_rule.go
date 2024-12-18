package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

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
	timeEngineRule := req.Do()
	return b.timeEngineRuleRepository.CreateTimeEngineRule(ctx, timeEngineRule)
}

// UpdateTimeEngineRule 更新时间引擎规则
func (b *TimeEngineRuleBiz) UpdateTimeEngineRule(ctx context.Context, req *bo.UpdateTimeEngineRuleRequest) error {
	timeEngineRule := req.Do()
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
