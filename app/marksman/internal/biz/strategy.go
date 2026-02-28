package biz

import (
	"context"

	"github.com/aide-family/magicbox/merr"
	"github.com/bwmarrin/snowflake"
	klog "github.com/go-kratos/kratos/v2/log"

	"github.com/aide-family/marksman/internal/biz/bo"
	"github.com/aide-family/marksman/internal/biz/repository"
)

func NewStrategy(
	strategyGroupRepo repository.StrategyGroup,
	strategyRepo repository.Strategy,
	helper *klog.Helper,
) *StrategyBiz {
	return &StrategyBiz{
		strategyGroupRepo: strategyGroupRepo,
		strategyRepo:      strategyRepo,
		helper:            klog.NewHelper(klog.With(helper.Logger(), "biz", "strategy")),
	}
}

type StrategyBiz struct {
	helper            *klog.Helper
	strategyGroupRepo repository.StrategyGroup
	strategyRepo      repository.Strategy
}

func (b *StrategyBiz) CreateStrategyGroup(ctx context.Context, req *bo.CreateStrategyGroupBo) error {
	if err := b.strategyGroupRepo.CreateStrategyGroup(ctx, req); err != nil {
		b.helper.Errorw("msg", "create strategy group failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create strategy group failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) UpdateStrategyGroup(ctx context.Context, req *bo.UpdateStrategyGroupBo) error {
	if err := b.strategyGroupRepo.UpdateStrategyGroup(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy group %d not found", req.UID.Int64())
		}
		b.helper.Errorw("msg", "update strategy group failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy group failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) UpdateStrategyGroupStatus(ctx context.Context, req *bo.UpdateStrategyGroupStatusBo) error {
	if err := b.strategyGroupRepo.UpdateStrategyGroupStatus(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy group %d not found", req.UID.Int64())
		}
		b.helper.Errorw("msg", "update strategy group status failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy group status failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) DeleteStrategyGroup(ctx context.Context, uid snowflake.ID) error {
	if err := b.strategyGroupRepo.DeleteStrategyGroup(ctx, uid); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy group %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "delete strategy group failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete strategy group failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) GetStrategyGroup(ctx context.Context, uid snowflake.ID) (*bo.StrategyGroupItemBo, error) {
	item, err := b.strategyGroupRepo.GetStrategyGroup(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("strategy group %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "get strategy group failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get strategy group failed").WithCause(err)
	}
	return item, nil
}

func (b *StrategyBiz) ListStrategyGroup(ctx context.Context, req *bo.ListStrategyGroupBo) (*bo.PageResponseBo[*bo.StrategyGroupItemBo], error) {
	result, err := b.strategyGroupRepo.ListStrategyGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "list strategy group failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list strategy group failed").WithCause(err)
	}
	return result, nil
}

func (b *StrategyBiz) SelectStrategyGroup(ctx context.Context, req *bo.SelectStrategyGroupBo) (*bo.SelectStrategyGroupBoResult, error) {
	result, err := b.strategyGroupRepo.SelectStrategyGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "select strategy group failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("select strategy group failed").WithCause(err)
	}
	return result, nil
}

func (b *StrategyBiz) StrategyGroupBindReceivers(ctx context.Context, req *bo.StrategyGroupBindReceiversBo) error {
	if err := b.strategyGroupRepo.StrategyGroupBindReceivers(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy group %d not found", req.StrategyGroupUID.Int64())
		}
		b.helper.Errorw("msg", "strategy group bind receivers failed", "error", err, "req", req)
		return merr.ErrorInternalServer("strategy group bind receivers failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) CreateStrategy(ctx context.Context, req *bo.CreateStrategyBo) error {
	if err := b.strategyRepo.CreateStrategy(ctx, req); err != nil {
		b.helper.Errorw("msg", "create strategy failed", "error", err, "req", req)
		return merr.ErrorInternalServer("create strategy failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) UpdateStrategy(ctx context.Context, req *bo.UpdateStrategyBo) error {
	if err := b.strategyRepo.UpdateStrategy(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy %d not found", req.UID.Int64())
		}
		b.helper.Errorw("msg", "update strategy failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) UpdateStrategyStatus(ctx context.Context, req *bo.UpdateStrategyStatusBo) error {
	if err := b.strategyRepo.UpdateStrategyStatus(ctx, req); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy %d not found", req.UID.Int64())
		}
		b.helper.Errorw("msg", "update strategy status failed", "error", err, "req", req)
		return merr.ErrorInternalServer("update strategy status failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) DeleteStrategy(ctx context.Context, uid snowflake.ID) error {
	if err := b.strategyRepo.DeleteStrategy(ctx, uid); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "delete strategy failed", "error", err, "uid", uid)
		return merr.ErrorInternalServer("delete strategy failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) GetStrategy(ctx context.Context, uid snowflake.ID) (*bo.StrategyItemBo, error) {
	item, err := b.strategyRepo.GetStrategy(ctx, uid)
	if err != nil {
		if merr.IsNotFound(err) {
			return nil, merr.ErrorNotFound("strategy %d not found", uid.Int64())
		}
		b.helper.Errorw("msg", "get strategy failed", "error", err, "uid", uid)
		return nil, merr.ErrorInternalServer("get strategy failed").WithCause(err)
	}
	return item, nil
}

func (b *StrategyBiz) ListStrategy(ctx context.Context, req *bo.ListStrategyBo) (*bo.PageResponseBo[*bo.StrategyItemBo], error) {
	result, err := b.strategyRepo.ListStrategy(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "list strategy failed", "error", err, "req", req)
		return nil, merr.ErrorInternalServer("list strategy failed").WithCause(err)
	}
	return result, nil
}
