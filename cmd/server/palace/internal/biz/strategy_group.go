package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
)

// NewStrategyGroupBiz 创建策略分组业务
func NewStrategyGroupBiz(strategy repository.StrategyGroup) *StrategyGroupBiz {
	return &StrategyGroupBiz{
		strategyRepo: strategy,
	}
}

// NewStrategyCountBiz 创建策略计数业务
func NewStrategyCountBiz(strategyCount repository.StrategyCountRepo) *StrategyCountBiz {
	return &StrategyCountBiz{
		strategyCountRepo: strategyCount,
	}
}

type (
	// StrategyGroupBiz 策略分组业务
	StrategyGroupBiz struct {
		strategyRepo repository.StrategyGroup
	}
	// StrategyCountBiz 策略计数业务
	StrategyCountBiz struct {
		strategyCountRepo repository.StrategyCountRepo
	}
)

// CreateStrategyGroup 创建策略分组
func (s *StrategyGroupBiz) CreateStrategyGroup(ctx context.Context, params *bo.CreateStrategyGroupParams) (*bizmodel.StrategyGroup, error) {
	// 查询策略分组是否已经存在
	strategyGroup, err := s.strategyRepo.CreateStrategyGroup(ctx, params)
	if !types.IsNil(err) {
		return nil, err
	}
	return strategyGroup, nil
}

// UpdateStrategyGroup 更新策略分组
func (s *StrategyGroupBiz) UpdateStrategyGroup(ctx context.Context, params *bo.UpdateStrategyGroupParams) error {
	return s.strategyRepo.UpdateStrategyGroup(ctx, params)
}

// GetStrategyGroupDetail 获取策略分组详情
func (s *StrategyGroupBiz) GetStrategyGroupDetail(ctx context.Context, groupID uint32) (*bizmodel.StrategyGroup, error) {
	strategyGroup, err := s.strategyRepo.GetStrategyGroup(ctx, groupID)
	if !types.IsNil(err) {
		return nil, err
	}
	return strategyGroup, nil
}

// DeleteStrategyGroup 删除策略分组
func (s *StrategyGroupBiz) DeleteStrategyGroup(ctx context.Context, params *bo.DelStrategyGroupParams) error {
	return s.strategyRepo.DeleteStrategyGroup(ctx, params)
}

// UpdateStatus 更新策略分组状态
func (s *StrategyGroupBiz) UpdateStatus(ctx context.Context, params *bo.UpdateStrategyGroupStatusParams) error {
	return s.strategyRepo.UpdateStatus(ctx, params)
}

// ListPage 分页查询策略分组
func (s *StrategyGroupBiz) ListPage(ctx context.Context, params *bo.QueryStrategyGroupListParams) ([]*bizmodel.StrategyGroup, error) {
	strategyGroups, err := s.strategyRepo.StrategyGroupPage(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return strategyGroups, err
}

// StrategyCount 策略分组关联策略总数
func (b *StrategyCountBiz) StrategyCount(ctx context.Context, params *bo.GetStrategyCountParams) []*bo.StrategyCountModel {
	strategyCount, err := b.strategyCountRepo.FindStrategyCount(ctx, params)
	if !types.IsNil(err) {
		return nil
	}
	return strategyCount
}
