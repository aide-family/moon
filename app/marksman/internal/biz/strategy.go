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

func NewStrategy(
	transaction repository.Transaction,
	strategyGroupRepo repository.StrategyGroup,
	strategyRepo repository.Strategy,
	strategyMetricRepo repository.StrategyMetric,
	helper *klog.Helper,
) *StrategyBiz {
	b := &StrategyBiz{
		transaction:        transaction,
		strategyGroupRepo:  strategyGroupRepo,
		strategyRepo:       strategyRepo,
		strategyMetricRepo: strategyMetricRepo,
		helper:             klog.NewHelper(klog.With(helper.Logger(), "biz", "strategy")),
	}
	b.deleteStrategyFunc = safety.NewMap(map[enum.DatasourceType]func(ctx context.Context, strategyUID snowflake.ID) error{
		enum.DatasourceType_METRICS: b.deleteStrategyMetric,
		enum.DatasourceType_LOGS:    b.deleteStrategyLogs,
		enum.DatasourceType_TRACE:   b.deleteStrategyTrace,
	})
	b.typeDetailCheckers = safety.NewMap(map[enum.DatasourceType]func(ctx context.Context, strategyUID snowflake.ID) (bool, error){
		enum.DatasourceType_METRICS: b.hasMetricDetail,
		// enum.DatasourceType_LOGS:  b.hasLogsDetail,   // add when logs detail exists
		// enum.DatasourceType_TRACE: b.hasTraceDetail, // add when trace detail exists
	})
	return b
}

type StrategyBiz struct {
	transaction        repository.Transaction
	helper             *klog.Helper
	strategyGroupRepo  repository.StrategyGroup
	strategyRepo       repository.Strategy
	strategyMetricRepo repository.StrategyMetric
	deleteStrategyFunc *safety.Map[enum.DatasourceType, func(ctx context.Context, strategyUID snowflake.ID) error]
	typeDetailCheckers *safety.Map[enum.DatasourceType, func(ctx context.Context, strategyUID snowflake.ID) (bool, error)]
}

func (b *StrategyBiz) CreateStrategyGroup(ctx context.Context, req *bo.CreateStrategyGroupBo) (snowflake.ID, error) {
	uid, err := b.strategyGroupRepo.CreateStrategyGroup(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "create strategy group failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("create strategy group failed").WithCause(err)
	}
	return uid, nil
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

func (b *StrategyBiz) CreateStrategy(ctx context.Context, req *bo.CreateStrategyBo) (snowflake.ID, error) {
	if _, err := b.strategyGroupRepo.GetStrategyGroup(ctx, req.StrategyGroupUID); err != nil {
		if merr.IsNotFound(err) {
			return 0, merr.ErrorNotFound("strategy group not found or invalid, please select a valid strategy group from the list")
		}
		b.helper.Errorw("msg", "get strategy group failed", "error", err, "strategyGroupUID", req.StrategyGroupUID)
		return 0, merr.ErrorInternalServer("create strategy failed").WithCause(err)
	}
	uid, err := b.strategyRepo.CreateStrategy(ctx, req)
	if err != nil {
		b.helper.Errorw("msg", "create strategy failed", "error", err, "req", req)
		return 0, merr.ErrorInternalServer("create strategy failed").WithCause(err)
	}
	return uid, nil
}

func (b *StrategyBiz) UpdateStrategy(ctx context.Context, req *bo.UpdateStrategyBo) error {
	if _, err := b.strategyGroupRepo.GetStrategyGroup(ctx, req.StrategyGroupUID); err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy group not found or invalid, please select a valid strategy group from the list")
		}
		b.helper.Errorw("msg", "get strategy group failed", "error", err, "strategyGroupUID", req.StrategyGroupUID)
		return merr.ErrorInternalServer("update strategy failed").WithCause(err)
	}
	current, err := b.strategyRepo.GetStrategy(ctx, req.UID)
	if err != nil {
		if merr.IsNotFound(err) {
			return merr.ErrorNotFound("strategy %d not found", req.UID.Int64())
		}
		b.helper.Errorw("msg", "get strategy failed", "error", err, "uid", req.UID)
		return merr.ErrorInternalServer("update strategy failed").WithCause(err)
	}
	if req.Type != current.Type || req.Driver != current.Driver {
		if checker, ok := b.typeDetailCheckers.Get(current.Type); ok && checker != nil {
			hasDetail, err := checker(ctx, req.UID)
			if err != nil {
				b.helper.Errorw("msg", "check strategy type detail failed", "error", err, "uid", req.UID, "type", current.Type)
				return merr.ErrorInternalServer("update strategy failed").WithCause(err)
			}
			if hasDetail {
				return merr.ErrorParams("cannot change type or driver when strategy metric or level data already exists, please delete them first")
			}
		}
	}
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

func (b *StrategyBiz) DeleteStrategy(ctx context.Context, uid snowflake.ID) error {
	strategyInfo, err := b.GetStrategy(ctx, uid)
	if err != nil {
		return err
	}
	return b.transaction.Transaction(ctx, func(ctx context.Context) error {
		if deleteStrategyFunc, ok := b.deleteStrategyFunc.Get(strategyInfo.Type); ok {
			if err := deleteStrategyFunc(ctx, uid); err != nil {
				return err
			}
		}
		if err := b.strategyRepo.DeleteStrategy(ctx, uid); err != nil {
			if merr.IsNotFound(err) {
				return merr.ErrorNotFound("strategy %d not found", uid.Int64())
			}
			b.helper.Errorw("msg", "delete strategy failed", "error", err, "uid", uid)
			return merr.ErrorInternalServer("delete strategy failed").WithCause(err)
		}
		return nil
	})
}

func (b *StrategyBiz) hasMetricDetail(ctx context.Context, strategyUID snowflake.ID) (bool, error) {
	hasMetric, err := b.strategyMetricRepo.HasStrategyMetricData(ctx, strategyUID)
	if err != nil {
		return false, err
	}
	if hasMetric {
		return true, nil
	}
	return b.strategyMetricRepo.HasStrategyMetricLevelData(ctx, strategyUID)
}

func (b *StrategyBiz) deleteStrategyMetric(ctx context.Context, strategyUID snowflake.ID) error {
	if err := b.strategyMetricRepo.DeleteStrategyMetricReceiversByStrategyUID(ctx, strategyUID); err != nil {
		b.helper.Errorw("msg", "delete strategy metric receivers failed", "error", err, "strategyUID", strategyUID)
		return merr.ErrorInternalServer("delete strategy related data failed").WithCause(err)
	}
	if err := b.strategyMetricRepo.DeleteStrategyMetricLevelsByStrategyUID(ctx, strategyUID); err != nil {
		b.helper.Errorw("msg", "delete strategy metric levels failed", "error", err, "strategyUID", strategyUID)
		return merr.ErrorInternalServer("delete strategy related data failed").WithCause(err)
	}
	if err := b.strategyMetricRepo.DeleteStrategyMetricByStrategyUID(ctx, strategyUID); err != nil {
		b.helper.Errorw("msg", "delete strategy metric failed", "error", err, "strategyUID", strategyUID)
		return merr.ErrorInternalServer("delete strategy related data failed").WithCause(err)
	}
	return nil
}

func (b *StrategyBiz) deleteStrategyLogs(ctx context.Context, strategyUID snowflake.ID) error {
	return nil
}

func (b *StrategyBiz) deleteStrategyTrace(ctx context.Context, strategyUID snowflake.ID) error {
	return nil
}
