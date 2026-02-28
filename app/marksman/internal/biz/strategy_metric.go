package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewStrategyMetric(
	strategyMetricRepo repository.StrategyMetric,
	helper *klog.Helper,
) *StrategyMetricBiz {
	return &StrategyMetricBiz{
		strategyMetricRepo: strategyMetricRepo,
		helper:             klog.NewHelper(klog.With(helper.Logger(), "biz", "strategy_metric")),
	}
}

type StrategyMetricBiz struct {
	helper             *klog.Helper
	strategyMetricRepo repository.StrategyMetric
}

func (b *StrategyMetricBiz) SaveStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error {
	_, err := b.strategyMetricRepo.GetStrategyMetric(ctx, req.StrategyUID)
	if err != nil {
		if !merr.IsNotFound(err) {
			b.helper.Errorw("msg", "get strategy metric failed", "error", err, "strategyUID", req.StrategyUID)
			return merr.ErrorInternalServer("get strategy metric failed").WithCause(err)
		}
		if err := b.strategyMetricRepo.CreateStrategyMetric(ctx, req); err != nil {
			b.helper.Errorw("msg", "create strategy metric failed", "error", err, "req", req)
			return merr.ErrorInternalServer("create strategy metric failed").WithCause(err)
		}
		return nil
	}
	if err := b.strategyMetricRepo.UpdateStrategyMetric(ctx, req); err != nil {
		b.helper.Errorw("msg", "update strategy metric failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy metric failed").WithCause(err)
	}
	return nil
}

func (b *StrategyMetricBiz) GetStrategyMetric(ctx context.Context, strategyUID snowflake.ID) (*bo.StrategyMetricItemBo, error) {
	item, err := b.strategyMetricRepo.GetStrategyMetric(ctx, strategyUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("strategy metric %d not found", strategyUID.Int64())
		}
		b.helper.Errorw("msg", "get strategy metric failed", "error", err, "strategyUID", strategyUID)
		return nil, merr.ErrorInternalServer("get strategy metric failed").WithCause(err)
	}
	return item, nil
}

func (b *StrategyMetricBiz) SaveStrategyMetricLevel(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) error {
	_, err := b.strategyMetricRepo.GetStrategyMetricLevel(ctx, req.StrategyUID, req.LevelUID)
	if err != nil {
		if !merr.IsNotFound(err) {
			b.helper.Errorw("msg", "get strategy metric level failed", "error", err, "strategyUID", req.StrategyUID, "levelUID", req.LevelUID)
			return merr.ErrorInternalServer("get strategy metric level failed").WithCause(err)
		}
		if err := b.strategyMetricRepo.CreateStrategyMetricLevel(ctx, req); err != nil {
			b.helper.Errorw("msg", "create strategy metric level failed", "error", err, "req", req)
			return merr.ErrorInternalServer("create strategy metric level failed").WithCause(err)
		}
		return nil
	}
	if err := b.strategyMetricRepo.UpdateStrategyMetricLevel(ctx, req); err != nil {
		b.helper.Errorw("msg", "update strategy metric level failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy metric level failed").WithCause(err)
	}
	return nil
}

func (b *StrategyMetricBiz) UpdateStrategyMetricLevelStatus(ctx context.Context, req *bo.UpdateStrategyMetricLevelStatusBo) error {
	if err := b.strategyMetricRepo.UpdateStrategyMetricLevelStatus(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy metric level not found")
		}
		b.helper.Errorw("msg", "update strategy metric level status failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy metric level status failed").WithCause(err)
	}
	return nil
}

func (b *StrategyMetricBiz) DeleteStrategyMetricLevel(ctx context.Context, uid snowflake.ID, strategyUID snowflake.ID) error {
	if err := b.strategyMetricRepo.DeleteStrategyMetricLevel(ctx, uid, strategyUID); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy metric level not found")
		}
		b.helper.Errorw("msg", "delete strategy metric level failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete strategy metric level failed").WithCause(err)
	}
	return nil
}

func (b *StrategyMetricBiz) GetStrategyMetricLevel(ctx context.Context, uid snowflake.ID, strategyUID snowflake.ID) (*bo.StrategyMetricLevelItemBo, error) {
	item, err := b.strategyMetricRepo.GetStrategyMetricLevel(ctx, uid, strategyUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("strategy metric level not found")
		}
		b.helper.Errorw("msg", "get strategy metric level failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get strategy metric level failed").WithCause(err)
	}
	return item, nil
}

func (b *StrategyMetricBiz) StrategyMetricBindReceivers(ctx context.Context, req *bo.StrategyMetricBindReceiversBo) error {
	if err := b.strategyMetricRepo.StrategyMetricBindReceivers(ctx, req); err != nil {
		b.helper.Errorw("msg", "strategy metric bind receivers failed", "error", err, "req", req)
		return merr.ErrorInternalServer("strategy metric bind receivers failed").WithCause(err)
	}
	return nil
}
