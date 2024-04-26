package biz

import (
	"context"
	"sort"
	"strings"

	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/gorm"
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
	l.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		ModuleName: vobj.ModuleStrategyGroup,
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
			ModuleName: vobj.ModuleStrategyGroup,
			ModuleId:   strategyGroup.Id,
			Content:    strategyGroup.String(),
			Title:      "批量创建策略组",
		}
	})
	l.logX.CreateSysLog(ctx, vobj.ActionCreate, list...)
	return strategyGroups, nil
}

func (l *StrategyGroupBiz) UpdateById(ctx context.Context, strategyGroup *bo.StrategyGroupBO) (*bo.StrategyGroupBO, error) {
	// 查询
	oldStrategyGroupBO, err := l.GetById(ctx, strategyGroup.Id)
	if err != nil {
		return nil, err
	}
	strategyGroup.Categories = slices.To(strategyGroup.GetCategoryIds(), func(category uint32) *bo.DictBO { return &bo.DictBO{Id: category} })
	strategyGroupBO, err := l.strategyGroupRepo.UpdateById(ctx, strategyGroup.Id, strategyGroup)
	if err != nil {
		return nil, err
	}
	l.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleStrategyGroup,
		ModuleId:   strategyGroup.Id,
		Content:    bo.NewChangeLogBo(oldStrategyGroupBO, strategyGroupBO).String(),
		Title:      "更新策略组",
	})
	return strategyGroupBO, nil
}

func (l *StrategyGroupBiz) BatchUpdateStatus(ctx context.Context, status vobj.Status, ids []uint32) error {
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
			ModuleName: vobj.ModuleStrategyGroup,
			ModuleId:   strategyGroupBO.Id,
			Content:    bo.NewChangeLogBo(strategyGroupBO.Status.String(), status.String()).String(),
			Title:      "批量更新策略组状态",
		}
	})
	l.logX.CreateSysLog(ctx, vobj.ActionUpdate, list...)
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
			ModuleName: vobj.ModuleStrategyGroup,
			ModuleId:   strategyGroupBO.Id,
			Content:    strategyGroupBO.String(),
			Title:      "批量删除策略组",
		}
	})
	l.logX.CreateSysLog(ctx, vobj.ActionDelete, list...)
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
		do.StrategyGroupWhereCategories(req.CategoryIds...),
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

// ListAllGroupDetail 获取所有策略组详情
func (l *StrategyGroupBiz) ListAllGroupDetail(ctx context.Context, params *bo.ListAllGroupDetailParams) ([]*bo.StrategyGroupBO, error) {
	list := make([]*bo.StrategyGroupBO, 0)
	wheres := []basescopes.ScopeMethod{
		basescopes.StatusEQ(vobj.StatusEnabled),
		func(db *gorm.DB) *gorm.DB {
			return db.Preload(do.PromGroupPreloadFieldPromStrategies, basescopes.StatusEQ(vobj.StatusEnabled))
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Preload(strings.Join([]string{
				do.PromGroupPreloadFieldPromStrategies,
				do.PromStrategyPreloadFieldEndpoint}, "."), basescopes.StatusEQ(vobj.StatusEnabled))
		},
		func(db *gorm.DB) *gorm.DB {
			return db.Preload(strings.Join([]string{
				do.PromGroupPreloadFieldPromStrategies,
				do.PromStrategyPreloadFieldAlertLevel}, "."), basescopes.StatusEQ(vobj.StatusEnabled))
		},
	}

	defaultId := uint32(0)
	if len(params.GroupIds) > 0 {
		wheres = append(wheres, basescopes.InIds(params.GroupIds...))
		ids := params.GroupIds
		// 排序
		sort.Slice(ids, func(i, j int) bool {
			return ids[i] > ids[j]
		})
		defaultId = ids[0] - 1
	}

	for {
		strategyGroupBOS, err := l.strategyGroupRepo.ListAllLimit(ctx, 1000, append(wheres, basescopes.IdGT(defaultId))...)
		if err != nil {
			l.log.Errorf("ListAllGroupDetail error: %v", err)
			return nil, err
		}
		list = append(list, strategyGroupBOS...)
		if len(strategyGroupBOS) == 0 || len(strategyGroupBOS) < 1000 {
			break
		}
		defaultId = strategyGroupBOS[len(strategyGroupBOS)-1].Id
	}

	return list, nil
}
