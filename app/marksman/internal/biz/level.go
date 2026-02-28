package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewLevel(
	levelRepo repository.Level,
	helper *klog.Helper,
) *LevelBiz {
	return &LevelBiz{
		levelRepo: levelRepo,
		helper:    klog.NewHelper(klog.With(helper.Logger(), "biz", "level")),
	}
}

type LevelBiz struct {
	helper    *klog.Helper
	levelRepo repository.Level
}

func (l *LevelBiz) CreateLevel(ctx context.Context, req *bo.CreateLevelBo) error {
	if err := l.levelRepo.CreateLevel(ctx, req); err != nil {
		l.helper.Errorw("msg", "create level failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create level failed").WithCause(err)
	}
	return nil
}

func (l *LevelBiz) UpdateLevel(ctx context.Context, req *bo.UpdateLevelBo) error {
	if err := l.levelRepo.UpdateLevel(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("level %d not found", req.UID.Int64())
		}
		l.helper.Errorw("msg", "update level failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update level failed").WithCause(err)
	}
	return nil
}

func (l *LevelBiz) UpdateLevelStatus(ctx context.Context, req *bo.UpdateLevelStatusBo) error {
	if err := l.levelRepo.UpdateLevelStatus(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("level %d not found", req.UID.Int64())
		}
		l.helper.Errorw("msg", "update level status failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update level status failed").WithCause(err)
	}
	return nil
}

func (l *LevelBiz) DeleteLevel(ctx context.Context, uid snowflake.ID) error {
	if err := l.levelRepo.DeleteLevel(ctx, uid); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("level %d not found", uid.Int64())
		}
		l.helper.Errorw("msg", "delete level failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete level failed").WithCause(err)
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
