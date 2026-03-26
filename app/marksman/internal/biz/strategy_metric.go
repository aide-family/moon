package biz

import (
	"context"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewStrategyMetric(
	strategyRepo repository.Strategy,
	strategyMetricRepo repository.StrategyMetric,
	levelRepo repository.Level,
	evaluateBiz *Evaluate,
	helper *klog.Helper,
) *StrategyMetricBiz {
	return &StrategyMetricBiz{
		strategyRepo:       strategyRepo,
		strategyMetricRepo: strategyMetricRepo,
		levelRepo:          levelRepo,
		evaluateBiz:        evaluateBiz,
		helper:             klog.NewHelper(klog.With(helper.Logger(), "biz", "strategy_metric")),
	}
}

type StrategyMetricBiz struct {
	helper             *klog.Helper
	strategyRepo       repository.Strategy
	strategyMetricRepo repository.StrategyMetric
	levelRepo          repository.Level
	evaluateBiz        *Evaluate
}

// validateStrategyExistsAndIsMetrics returns the strategy if it exists and is METRICS type; otherwise returns a user-friendly error.
func (b *StrategyMetricBiz) validateStrategyExistsAndIsMetrics(ctx context.Context, strategyUID snowflake.ID) (*bo.StrategyItemBo, error) {
	strategy, err := b.strategyRepo.GetStrategy(ctx, strategyUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("strategy not found or deleted, please check and try again")
		}
		b.helper.Errorw("msg", "get strategy failed", "error", err, "strategyUID", strategyUID)
		return nil, merr.ErrorInternalServer("get strategy failed").WithCause(err)
	}
	if strategy.Type != enum.DatasourceType_METRICS {
		return nil, merr.ErrorParams("only supports configuring metrics for METRICS type strategies, current strategy type does not match, please select a METRICS type strategy")
	}
	return strategy, nil
}

func (b *StrategyMetricBiz) SaveStrategyMetric(ctx context.Context, req *bo.SaveStrategyMetricBo) error {
	if _, err := b.validateStrategyExistsAndIsMetrics(ctx, req.StrategyUID); err != nil {
		return err
	}
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
		b.evaluateBiz.SyncByStrategyUID(ctx, req.StrategyUID)
		return nil
	}
	if err := b.strategyMetricRepo.UpdateStrategyMetric(ctx, req); err != nil {
		b.helper.Errorw("msg", "update strategy metric failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy metric failed").WithCause(err)
	}
	b.evaluateBiz.SyncByStrategyUID(ctx, req.StrategyUID)
	return nil
}

func (b *StrategyMetricBiz) GetStrategyMetric(ctx context.Context, strategyUID snowflake.ID) (*bo.StrategyMetricItemBo, error) {
	strategyBo, err := b.strategyRepo.GetStrategy(ctx, strategyUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("strategy not found")
		}
		b.helper.Errorw("msg", "get strategy failed", "error", err, "strategyUID", strategyUID)
		return nil, merr.ErrorInternalServer("get strategy failed").WithCause(err)
	}
	item, err := b.strategyMetricRepo.GetStrategyMetric(ctx, strategyUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return &bo.StrategyMetricItemBo{
				StrategyUID: strategyUID,
				Strategy:    strategyBo,
			}, nil
		}
		b.helper.Errorw("msg", "get strategy metric failed", "error", err, "strategyUID", strategyUID)
		return nil, merr.ErrorInternalServer("get strategy metric failed").WithCause(err)
	}
	return item, nil
}

// validateLevelExistsAndEnabled returns the level if it exists and is enabled; otherwise returns a user-friendly error.
func (b *StrategyMetricBiz) validateLevelExistsAndEnabled(ctx context.Context, levelUID snowflake.ID) (*bo.LevelItemBo, error) {
	level, err := b.levelRepo.GetLevel(ctx, levelUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("level not found or deleted, please select a valid level")
		}
		b.helper.Errorw("msg", "get level failed", "error", err, "levelUID", levelUID)
		return nil, merr.ErrorInternalServer("get level failed").WithCause(err)
	}
	if level.Type != enum.LevelType_LEVEL_TYPE_ALERT {
		return nil, merr.ErrorParams("the selected level is not an ALERT level, please select a valid level")
	}
	if level.Status != enum.GlobalStatus_ENABLED {
		return nil, merr.ErrorParams("the selected level has been disabled, please select a new one")
	}
	return level, nil
}

func (b *StrategyMetricBiz) SaveStrategyMetricLevel(ctx context.Context, req *bo.SaveStrategyMetricLevelBo) error {
	if _, err := b.validateStrategyExistsAndIsMetrics(ctx, req.StrategyUID); err != nil {
		return err
	}
	if _, err := b.validateLevelExistsAndEnabled(ctx, req.LevelUID); err != nil {
		return err
	}
	_, err := b.strategyMetricRepo.GetStrategyMetricLevelByStrategyAndLevel(ctx, req.StrategyUID, req.LevelUID)
	if err != nil {
		if !merr.IsNotFound(err) {
			b.helper.Errorw("msg", "get strategy metric level failed", "error", err, "strategyUID", req.StrategyUID, "levelUID", req.LevelUID)
			return merr.ErrorInternalServer("get strategy metric level failed").WithCause(err)
		}
		if err := b.strategyMetricRepo.CreateStrategyMetricLevel(ctx, req); err != nil {
			b.helper.Errorw("msg", "create strategy metric level failed", "error", err, "req", req)
			return merr.ErrorInternalServer("create strategy metric level failed").WithCause(err)
		}
		b.evaluateBiz.SyncByStrategyLevelUID(ctx, req.StrategyUID, req.LevelUID)
		return nil
	}
	if err := b.strategyMetricRepo.UpdateStrategyMetricLevel(ctx, req); err != nil {
		b.helper.Errorw("msg", "update strategy metric level failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy metric level failed").WithCause(err)
	}
	b.evaluateBiz.SyncByStrategyLevelUID(ctx, req.StrategyUID, req.LevelUID)
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
	if req.Status == enum.GlobalStatus_ENABLED {
		b.evaluateBiz.SyncByStrategyLevelUID(ctx, req.StrategyUID, req.LevelUID)
	} else {
		b.evaluateBiz.RemoveByStrategyLevelUID(ctx, req.StrategyUID, req.LevelUID)
	}
	return nil
}

func (b *StrategyMetricBiz) DeleteStrategyMetricLevel(ctx context.Context, levelUID snowflake.ID, strategyUID snowflake.ID) error {
	if err := b.strategyMetricRepo.DeleteStrategyMetricLevel(ctx, levelUID, strategyUID); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy metric level not found")
		}
		b.helper.Errorw("msg", "delete strategy metric level failed", "error", err, "levelUID", levelUID)
		return merr.ErrorInternalServer("delete strategy metric level failed").WithCause(err)
	}
	b.evaluateBiz.RemoveByStrategyLevelUID(ctx, strategyUID, levelUID)
	return nil
}

func (b *StrategyMetricBiz) GetStrategyMetricLevel(ctx context.Context, levelUID snowflake.ID, strategyUID snowflake.ID) (*bo.StrategyMetricLevelItemBo, error) {
	item, err := b.strategyMetricRepo.GetStrategyMetricLevelByStrategyAndLevel(ctx, strategyUID, levelUID)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("strategy metric level not found")
		}
		b.helper.Errorw("msg", "get strategy metric level failed", "error", err, "levelUID", levelUID)
		return nil, merr.ErrorInternalServer("get strategy metric level failed").WithCause(err)
	}
	return item, nil
}
