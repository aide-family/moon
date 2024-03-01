package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/api/perrors"

	"prometheus-manager/api"
	strategyPB "prometheus-manager/api/prom/strategy"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyBiz struct {
		log *log.Helper

		strategyRepo repository.StrategyRepo
		notifyRepo   repository.NotifyRepo
		logX         repository.SysLogRepo
	}
)

// NewStrategyBiz 创建策略业务对象
func NewStrategyBiz(strategyRepo repository.StrategyRepo, notifyRepo repository.NotifyRepo, logX repository.SysLogRepo, logger log.Logger) *StrategyBiz {
	return &StrategyBiz{
		log: log.NewHelper(log.With(logger, "module", "strategy")),

		strategyRepo: strategyRepo,
		notifyRepo:   notifyRepo,
		logX:         logX,
	}
}

// CreateStrategy 创建策略
func (b *StrategyBiz) CreateStrategy(ctx context.Context, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	newStrategyBO := strategyBO
	newStrategyBO.AlarmPages = slices.To(strategyBO.AlarmPageIds, func(id uint32) *bo.AlarmPageBO {
		return &bo.AlarmPageBO{Id: id}
	})
	newStrategyBO.Categories = slices.To(strategyBO.CategoryIds, func(id uint32) *bo.DictBO {
		return &bo.DictBO{Id: id}
	})
	strategyBO, err := b.strategyRepo.CreateStrategy(ctx, strategyBO)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vo.ActionCreate, &bo.SysLogBo{
		ModuleName: vo.ModuleStrategy,
		ModuleId:   strategyBO.Id,
		Content:    strategyBO.String(),
		Title:      "创建策略",
	})

	return strategyBO, nil
}

// UpdateStrategyById 更新策略
func (b *StrategyBiz) UpdateStrategyById(ctx context.Context, id uint32, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	// 查询
	oldData, err := b.GetStrategyById(ctx, id)
	if err != nil {
		return nil, err
	}
	newStrategyBO, err := b.strategyRepo.UpdateStrategyById(ctx, id, strategyBO)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vo.ActionUpdate, &bo.SysLogBo{
		ModuleName: vo.ModuleStrategy,
		ModuleId:   strategyBO.Id,
		Content:    bo.NewChangeLogBo(oldData, newStrategyBO).String(),
		Title:      "更新策略",
	})

	return strategyBO, nil
}

// BatchUpdateStrategyStatusByIds 批量更新策略状态
func (b *StrategyBiz) BatchUpdateStrategyStatusByIds(ctx context.Context, status vo.Status, ids []uint32) error {
	oldList, err := b.strategyRepo.List(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err = b.strategyRepo.BatchUpdateStrategyStatusByIds(ctx, status, ids); err != nil {
		return err
	}

	list := slices.To(oldList, func(old *bo.StrategyBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleStrategy,
			ModuleId:   old.Id,
			Content:    bo.NewChangeLogBo(old.Status.String(), status.String()).String(),
			Title:      "批量更新策略状态",
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, list...)
	return nil
}

// DeleteStrategyByIds 删除策略
func (b *StrategyBiz) DeleteStrategyByIds(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}
	// 查询
	oldList, err := b.strategyRepo.List(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err = b.strategyRepo.DeleteStrategyByIds(ctx, ids...); err != nil {
		return err
	}
	list := slices.To(oldList, func(old *bo.StrategyBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleStrategy,
			ModuleId:   old.Id,
			Content:    old.String(),
			Title:      "删除策略",
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionDelete, list...)
	return nil
}

// GetStrategyById 获取策略详情
func (b *StrategyBiz) GetStrategyById(ctx context.Context, id uint32) (*bo.StrategyBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.StrategyTablePreloadEndpoint,
		basescopes.StrategyTablePreloadAlarmPages,
		basescopes.StrategyTablePreloadCategories,
		basescopes.StrategyTablePreloadAlertLevel,
		basescopes.StrategyTablePreloadPromNotifies(),
		basescopes.StrategyTablePreloadPromNotifyUpgrade,
		basescopes.StrategyTablePreloadGroupInfo,
	}
	strategyBO, err := b.strategyRepo.GetStrategyById(ctx, id, wheres...)
	if err != nil {
		return nil, err
	}

	return strategyBO, nil
}

// ListStrategy 获取策略列表
func (b *StrategyBiz) ListStrategy(ctx context.Context, req *strategyPB.ListStrategyRequest) ([]*bo.StrategyBO, basescopes.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	scopes := []basescopes.ScopeMethod{
		basescopes.StrategyTableAlertLike(req.GetKeyword()),
		basescopes.StrategyTableGroupIdsEQ(req.GetGroupId()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
		basescopes.StrategyTablePreloadAlertLevel,
		basescopes.StrategyTablePreloadCategories,
		basescopes.StrategyTablePreloadEndpoint,
		basescopes.StrategyTablePreloadGroupInfo,
		basescopes.StrategyTablePreloadAlarmPages,
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
		basescopes.InIds(req.GetStrategyId()),
	}

	strategyBOs, err := b.strategyRepo.ListStrategy(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return strategyBOs, pgInfo, nil
}

// SelectStrategy 查询策略
func (b *StrategyBiz) SelectStrategy(ctx context.Context, req *strategyPB.SelectStrategyRequest) ([]*bo.StrategyBO, basescopes.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := basescopes.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	scopes := []basescopes.ScopeMethod{
		basescopes.StrategyTableAlertLike(req.GetKeyword()),
		basescopes.StatusEQ(vo.Status(api.Status_STATUS_ENABLED)),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
	}

	strategyBOs, err := b.strategyRepo.ListStrategy(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return strategyBOs, pgInfo, nil
}

// ExportStrategy 导出策略
func (b *StrategyBiz) ExportStrategy(ctx context.Context, req *strategyPB.ExportStrategyRequest) ([]*bo.StrategyBO, error) {
	strategyBOs, err := b.strategyRepo.ListStrategyByIds(ctx, req.GetIds())
	if err != nil {
		return nil, err
	}

	list := slices.To(strategyBOs, func(strategyBO *bo.StrategyBO) string {
		return strategyBO.String()
	})
	b.logX.CreateSysLog(ctx, vo.ActionExport, &bo.SysLogBo{
		ModuleName: vo.ModuleStrategy,
		ModuleId:   strategyBOs[0].Id,
		Content:    fmt.Sprintf(`{"list":[%s]}`, strings.Join(list, ",")),
		Title:      "导出策略",
	})

	return strategyBOs, nil
}

// GetStrategyWithNotifyObjectById 获取策略详情（包含通知对象）
func (b *StrategyBiz) GetStrategyWithNotifyObjectById(ctx context.Context, id uint32) (*bo.StrategyBO, error) {
	wheres := []basescopes.ScopeMethod{
		basescopes.StrategyTablePreloadPromNotifies(
			basescopes.NotifyTablePreloadKeyChatGroups,
			basescopes.NotifyTablePreloadKeyBeNotifyMembers,
		),
	}
	return b.strategyRepo.GetStrategyById(ctx, id, wheres...)
}

// BindStrategyNotifyObject 绑定策略的通知对象
func (b *StrategyBiz) BindStrategyNotifyObject(ctx context.Context, strategyId uint32, notifyIds []uint32) error {
	// 查询策略详情
	strategyBO, err := b.GetStrategyById(ctx, strategyId)
	if err != nil {
		return err
	}

	// 查询通知对象
	notifyBOs, err := b.notifyRepo.Find(ctx, basescopes.InIds(notifyIds...))
	if err != nil {
		return err
	}
	if len(notifyBOs) != len(notifyIds) {
		return perrors.ErrorNotFound("notify not found")
	}

	if err = b.strategyRepo.BindStrategyNotifyObject(ctx, strategyBO, notifyBOs); err != nil {
		return err
	}

	notifyStr := slices.To(notifyBOs, func(notifyBO *bo.NotifyBO) string {
		return notifyBO.String()
	})

	list := slices.To(notifyBOs, func(notifyBO *bo.NotifyBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vo.ModuleStrategy,
			ModuleId:   strategyBO.Id,
			Content:    fmt.Sprintf(`"notifies":[%s]`, strings.Join(notifyStr, ",")),
			Title:      "绑定策略通知对象",
		}
	})
	b.logX.CreateSysLog(ctx, vo.ActionUpdate, list...)
	return nil
}
