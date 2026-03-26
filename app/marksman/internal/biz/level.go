package biz

import (
	"context"

	"github.com/aide-family/magicbox/enum"
	"github.com/aide-family/magicbox/merr"
	"github.com/aide-family/magicbox/safety"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewLevel(
	levelRepo repository.Level,
	strategyMetricRepo repository.StrategyMetric,
	evaluateBiz *Evaluate,
	helper *klog.Helper,
) *LevelBiz {
	b := &LevelBiz{
		levelRepo:          levelRepo,
		strategyMetricRepo: strategyMetricRepo,
		evaluateBiz:        evaluateBiz,
		helper:             klog.NewHelper(klog.With(helper.Logger(), "biz", "level")),
	}
	b.LevelReferencedFuncs = safety.NewSlice([]func(ctx context.Context, levelUID snowflake.ID) (bool, error){
		b.strategyMetricRepo.LevelReferencedByStrategyMetricLevel,
	})
	return b
}

type LevelBiz struct {
	helper               *klog.Helper
	levelRepo            repository.Level
	strategyMetricRepo   repository.StrategyMetric
	evaluateBiz          *Evaluate
	LevelReferencedFuncs *safety.Slice[func(ctx context.Context, levelUID snowflake.ID) (bool, error)]
}

func (l *LevelBiz) CreateLevel(ctx context.Context, req *bo.CreateLevelBo) (snowflake.ID, error) {
	taken, err := l.levelRepo.LevelNameTaken(ctx, req.Name, 0)
	if err != nil {
		l.helper.Errorw("msg", "check level name taken failed", "error", err, "name", req.Name)
		return 0, merr.ErrorInternalServer("check level name failed").WithCause(err)
	}
	if taken {
		return 0, merr.ErrorParams("level name already exists, please use another name")
	}
	uid, err := l.levelRepo.CreateLevel(ctx, req)
	if err != nil {
		l.helper.Errorw("msg", "create level failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("create level failed").WithCause(err)
	}
	return uid, nil
}

func (l *LevelBiz) UpdateLevel(ctx context.Context, req *bo.UpdateLevelBo) error {
	current, err := l.levelRepo.GetLevel(ctx, req.UID)
	if err != nil {
		return err
	}
	taken, err := l.levelRepo.LevelNameTaken(ctx, req.Name, req.UID)
	if err != nil {
		l.helper.Errorw("msg", "check level name taken failed", "error", err, "name", req.Name)
		return merr.ErrorInternalServer("check level name failed").WithCause(err)
	}
	if taken {
		return merr.ErrorParams("level name already exists, please use another name")
	}
	if err := l.levelRepo.UpdateLevel(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("level %d not found", req.UID.Int64())
		}
		l.helper.Errorw("msg", "update level failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update level failed").WithCause(err)
	}

	switch {
	case current.Type == enum.LevelType_LEVEL_TYPE_ALERT && req.Type == enum.LevelType_LEVEL_TYPE_ALERT:
		l.evaluateBiz.SyncByLevelUID(ctx, req.UID)
	case current.Type == enum.LevelType_LEVEL_TYPE_ALERT && req.Type != enum.LevelType_LEVEL_TYPE_ALERT:
		// Switch from ALERT to DATASOURCE should remove evaluation jobs for this level.
		l.evaluateBiz.RemoveByLevelUID(ctx, req.UID)
	case current.Type != enum.LevelType_LEVEL_TYPE_ALERT && req.Type == enum.LevelType_LEVEL_TYPE_ALERT:
		// Switch from DATASOURCE to ALERT should load evaluation jobs for this level.
		l.evaluateBiz.SyncByLevelUID(ctx, req.UID)
	default:
		// Both before/after are DATASOURCE: evaluation jobs should not be affected.
	}
	return nil
}

func (l *LevelBiz) UpdateLevelStatus(ctx context.Context, req *bo.UpdateLevelStatusBo) error {
	current, err := l.levelRepo.GetLevel(ctx, req.UID)
	if err != nil {
		return err
	}
	if err := l.levelRepo.UpdateLevelStatus(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("level %d not found", req.UID.Int64())
		}
		l.helper.Errorw("msg", "update level status failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update level status failed").WithCause(err)
	}
	if current.Type == enum.LevelType_LEVEL_TYPE_ALERT {
		if req.Status == enum.GlobalStatus_ENABLED {
			l.evaluateBiz.SyncByLevelUID(ctx, req.UID)
		} else {
			l.evaluateBiz.RemoveByLevelUID(ctx, req.UID)
		}
	}
	return nil
}

func (l *LevelBiz) DeleteLevel(ctx context.Context, uid snowflake.ID) error {
	current, err := l.levelRepo.GetLevel(ctx, uid)
	if err != nil {
		return err
	}
	for _, f := range l.LevelReferencedFuncs.List() {
		referenced, err := f(ctx, uid)
		if err != nil {
			l.helper.Errorw("msg", "check level referenced failed", "error", err, "uid", uid)
			return merr.ErrorInternalServer("check level referenced failed").WithCause(err)
		}
		if referenced {
			return merr.ErrorParams("level is referenced and cannot be deleted, please remove the reference first")
		}
	}
	if err := l.levelRepo.DeleteLevel(ctx, uid); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("level %d not found", uid.Int64())
		}
		l.helper.Errorw("msg", "delete level failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete level failed").WithCause(err)
	}
	if current.Type == enum.LevelType_LEVEL_TYPE_ALERT {
		l.evaluateBiz.RemoveByLevelUID(ctx, uid)
	}
	return nil
}

func (l *LevelBiz) GetLevel(ctx context.Context, uid snowflake.ID) (*bo.LevelItemBo, error) {
	item, err := l.levelRepo.GetLevel(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("level %d not found", uid.Int64())
		}
		l.helper.Errorw("msg", "get level failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get level failed").WithCause(err)
	}
	return item, nil
}

func (l *LevelBiz) ListLevel(ctx context.Context, req *bo.ListLevelBo) (*bo.PageResponseBo[*bo.LevelItemBo], error) {
	result, err := l.levelRepo.ListLevel(ctx, req)
	if err != nil {
		l.helper.Errorw("msg", "list level failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list level failed").WithCause(err)
	}
	return result, nil
}

func (l *LevelBiz) SelectLevel(ctx context.Context, req *bo.SelectLevelBo) (*bo.SelectLevelBoResult, error) {
	result, err := l.levelRepo.SelectLevel(ctx, req)
	if err != nil {
		l.helper.Errorw("msg", "select level failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select level failed").WithCause(err)
	}
	return result, nil
}
