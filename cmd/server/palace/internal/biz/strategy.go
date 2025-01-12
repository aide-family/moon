package biz

import (
	"context"

	strategyapi "github.com/aide-family/moon/api/houyi/strategy"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewStrategyBiz 创建策略业务
func NewStrategyBiz(dictRepo repository.Strategy, strategyRPCRepo microrepository.Strategy) *StrategyBiz {
	return &StrategyBiz{
		strategyRepo:    dictRepo,
		strategyRPCRepo: strategyRPCRepo,
	}
}

// StrategyBiz 策略业务
type StrategyBiz struct {
	strategyRepo    repository.Strategy
	strategyRPCRepo microrepository.Strategy
}

// GetStrategy 获取策略
func (b *StrategyBiz) GetStrategy(ctx context.Context, strategyID uint32) (*bizmodel.Strategy, error) {
	strategy, err := b.strategyRepo.GetByID(ctx, strategyID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastStrategyNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return strategy, nil
}

// CreateStrategy 创建策略
func (b *StrategyBiz) CreateStrategy(ctx context.Context, param *bo.CreateStrategyParams) (*bizmodel.Strategy, error) {
	strategyDo, err := b.strategyRepo.CreateStrategy(ctx, param)
	if !types.IsNil(err) {
		return nil, err
	}
	return strategyDo, nil
}

// UpdateByID 更新策略
func (b *StrategyBiz) UpdateByID(ctx context.Context, param *bo.UpdateStrategyParams) error {
	return b.strategyRepo.UpdateByID(ctx, param)
}

// UpdateStatus 更新策略状态
func (b *StrategyBiz) UpdateStatus(ctx context.Context, param *bo.UpdateStrategyStatusParams) error {
	// 校验策略分组是否打开
	if param.Status.IsEnable() {
		if err := b.verifyStrategyStatus(ctx, param.Ids); err != nil {
			return err
		}
	}
	err := b.strategyRepo.UpdateStatus(ctx, param)
	if !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// DeleteByID 删除策略
func (b *StrategyBiz) DeleteByID(ctx context.Context, strategyID uint32) error {
	err := b.strategyRepo.DeleteByID(ctx, strategyID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastStrategyNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// StrategyPage 获取策略分页
func (b *StrategyBiz) StrategyPage(ctx context.Context, param *bo.QueryStrategyListParams) ([]*bizmodel.Strategy, error) {
	strategies, err := b.strategyRepo.FindByPage(ctx, param)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return strategies, nil
}

// CopyStrategy 复制策略
func (b *StrategyBiz) CopyStrategy(ctx context.Context, strategyID uint32) (*bizmodel.Strategy, error) {
	strategy, err := b.strategyRepo.CopyStrategy(ctx, strategyID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return strategy, merr.ErrorI18nToastStrategyNotFound(ctx)
		}
		return strategy, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return strategy, nil
}

// Eval 策略评估
func (b *StrategyBiz) Eval(ctx context.Context, strategy *bo.Strategy) (*bo.Alarm, error) {
	return b.strategyRepo.Eval(ctx, strategy)
}

// PushStrategy 推送策略
func (b *StrategyBiz) PushStrategy(ctx context.Context, strategies *strategyapi.PushStrategyRequest) error {
	return b.strategyRPCRepo.Push(ctx, strategies)
}

// 校验策略分组是否打开,未打开策略分组不允许打开策略
func (b *StrategyBiz) verifyStrategyStatus(ctx context.Context, ids []uint32) error {
	strategies, err := b.strategyRepo.GetStrategyByIds(ctx, ids)
	if err != nil {
		return err
	}
	for _, strategy := range strategies {
		if types.IsNil(strategy.Group) {
			continue
		}
		if strategy.Group.Status.IsDisable() {
			return merr.ErrorI18nAlertStrategyGroupNotEnable(ctx, strategy.Name, strategy.Group.Name)
		}
	}
	return nil
}

// SyncStrategy 同步策略
func (b *StrategyBiz) SyncStrategy(ctx context.Context, id uint32) error {
	return b.strategyRepo.Sync(ctx, id)
}

// GetStrategyCategories 获取策略分类下的策略ID
func (b *StrategyBiz) GetStrategyCategories(ctx context.Context, ids []uint32, strategyTypes []vobj.StrategyType) ([]*bizmodel.StrategyCategories, error) {
	categories, err := b.strategyRepo.GetStrategyCategories(ctx, ids, strategyTypes)
	if types.IsNotNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return categories, nil
}
