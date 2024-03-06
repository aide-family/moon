package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyGroupBiz struct {
		log *log.Helper

		strategyGroupRepo repository.StrategyGroupRepo
		logX              repository.SysLogRepo
	}
)

func NewStrategyGroupBiz(strategyGroupRepo repository.StrategyGroupRepo, logX repository.SysLogRepo, logger log.Logger) *StrategyGroupBiz {
	return &StrategyGroupBiz{
		log:               log.NewHelper(log.With(logger, "module", "biz.strategyGroup")),
		strategyGroupRepo: strategyGroupRepo,
		logX:              logX,
	}
}

func (l *StrategyGroupBiz) Create(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	strategyGroup, err := l.strategyGroupRepo.Create(ctx, strategyGroup)
	if err != nil {
		return nil, err
	}
	l.logX.CreateSysLog(ctx, vo.ActionCreate, &bo.SysLogBo{
		ModuleName: vo.ModuleStrategyGroup,
		ModuleId:   strategyGroup.Id,
		Content:    strategyGroup.String(),
		Title:      "创建策略组",
	})
	return strategyGroup, nil
}

// BatchCreate 批量创建
func (l *StrategyGroupBiz) BatchCreate(ctx context.Context, strategyGroups []*bo.StrategyGroupBO) ([]*bo.StrategyGroupBO, error) {
	strategyGroups, err := l.strategyGroupRepo.BatchCreate(ctx, strategyGroups)
	if err != nil {
		return nil, err
	}
	list := slices.To(strategyGroups, func(strategyGroup *bo.StrategyGroupBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleStrategyGroup,
			ModuleId:   strategyGroup.Id,
			Content:    strategyGroup.String(),
			Title:      "批量创建策略组",
		}
	})
	l.logX.CreateSysLog(ctx, vo.ActionCreate, list...)
	return strategyGroups, nil
}

func (l *StrategyGroupBiz) UpdateById(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	// 查询
	oldStrategyGroupBO, err := l.GetById(ctx, strategyGroup.Id)
	if err != nil {
		return nil, err
	}
	strategyGroupBO, err := l.strategyGroupRepo.UpdateById(ctx, strategyGroup.Id, strategyGroup)
	if err != nil {
		return nil, err
	}
	l.logX.CreateSysLog(ctx, vo.ActionUpdate, &bo.SysLogBo{
		ModuleName: vo.ModuleStrategyGroup,
		ModuleId:   strategyGroup.Id,
		Content:    bo.NewChangeLogBo(oldStrategyGroupBO, strategyGroupBO).String(),
		Title:      "更新策略组",
	})
	return strategyGroupBO, nil
}

func (l *StrategyGroupBiz) BatchUpdateStatus(ctx context.Context, status vo.Status, ids []uint32) error {
	// 查询
	oldList, err := l.strategyGroupRepo.GetByParams(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err = l.strategyGroupRepo.BatchUpdateStatus(ctx, status, ids); err != nil {
		return err
	}
	list := slices.To(oldList, func(strategyGroupBO *bo.StrategyGroupBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleStrategyGroup,
			ModuleId:   strategyGroupBO.Id,
			Content:    bo.NewChangeLogBo(strategyGroupBO.Status.String(), status.String()).String(),
			Title:      "批量更新策略组状态",
		}
	})
	l.logX.CreateSysLog(ctx, vo.ActionUpdate, list...)
	return nil
}

func (l *StrategyGroupBiz) DeleteByIds(ctx context.Context, ids ...uint32) error {
	// 查询
	oldList, err := l.strategyGroupRepo.GetByParams(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err = l.strategyGroupRepo.DeleteByIds(ctx, ids...); err != nil {
		return err
	}
	list := slices.To(oldList, func(strategyGroupBO *bo.StrategyGroupBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleStrategyGroup,
			ModuleId:   strategyGroupBO.Id,
			Content:    strategyGroupBO.String(),
			Title:      "批量删除策略组",
		}
	})
	l.logX.CreateSysLog(ctx, vo.ActionDelete, list...)
	return nil
}

func (l *StrategyGroupBiz) GetById(ctx context.Context, id uint32) (*bo.StrategyGroupBO, error) {
	strategyGroupBO, err := l.strategyGroupRepo.GetById(ctx, id)
	if err != nil {
		return nil, err
	}
	return strategyGroupBO, nil
}

func (l *StrategyGroupBiz) List(ctx context.Context, req *bo.ListGroupReq) ([]*bo.StrategyGroupBO, error) {
	scopes := []basescopes.ScopeMethod{
		basescopes.NameLike(req.Keyword),
		basescopes.StatusEQ(req.Status),
		do.StrategyGroupPreloadCategories(req.PreloadCategories),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
		basescopes.InIds(req.Ids...),
	}
	strategyGroupBoList, err := l.strategyGroupRepo.List(ctx, req.Page, scopes...)
	if err != nil {
		return nil, err
	}
	return strategyGroupBoList, nil
}

func (l *StrategyGroupBiz) ListAllLimit(ctx context.Context, limit int, scopes ...basescopes.ScopeMethod) ([]*bo.StrategyGroupBO, error) {
	return l.strategyGroupRepo.ListAllLimit(ctx, limit, scopes...)
}
