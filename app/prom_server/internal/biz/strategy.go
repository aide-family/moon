package biz

import (
	"context"
	"fmt"
	"strings"

	"github.com/aide-family/moon/api/perrors"
	"github.com/aide-family/moon/app/prom_server/internal/biz/bo"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do"
	"github.com/aide-family/moon/app/prom_server/internal/biz/do/basescopes"
	"github.com/aide-family/moon/app/prom_server/internal/biz/repository"
	"github.com/aide-family/moon/app/prom_server/internal/biz/vobj"
	"github.com/aide-family/moon/pkg/util/slices"
	"github.com/go-kratos/kratos/v2/log"
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
	newStrategyBO.AlarmPages = slices.To(strategyBO.AlarmPageIds, func(id uint32) *bo.DictBO {
		return &bo.DictBO{Id: id}
	})
	newStrategyBO.Categories = slices.To(strategyBO.CategoryIds, func(id uint32) *bo.DictBO {
		return &bo.DictBO{Id: id}
	})
	strategyBO, err := b.strategyRepo.CreateStrategy(ctx, strategyBO)
	if err != nil {
		return nil, err
	}

	b.logX.CreateSysLog(ctx, vobj.ActionCreate, &bo.SysLogBo{
		ModuleName: vobj.ModuleStrategy,
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

	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, &bo.SysLogBo{
		ModuleName: vobj.ModuleStrategy,
		ModuleId:   strategyBO.Id,
		Content:    bo.NewChangeLogBo(oldData, newStrategyBO).String(),
		Title:      "更新策略",
	})

	return strategyBO, nil
}

// BatchUpdateStrategyStatusByIds 批量更新策略状态
func (b *StrategyBiz) BatchUpdateStrategyStatusByIds(ctx context.Context, status vobj.Status, ids []uint32) error {
	oldList, err := b.strategyRepo.List(ctx, basescopes.InIds(ids...))
	if err != nil {
		return err
	}
	if err = b.strategyRepo.BatchUpdateStrategyStatusByIds(ctx, status, ids); err != nil {
		return err
	}

	list := slices.To(oldList, func(old *bo.StrategyBO) *bo.SysLogBo {
		return &bo.SysLogBo{
			ModuleName: vobj.ModuleStrategy,
			ModuleId:   old.Id,
			Content:    bo.NewChangeLogBo(old.Status.String(), status.String()).String(),
			Title:      "批量更新策略状态",
		}
	})
	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, list...)
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
			ModuleName: vobj.ModuleStrategy,
			ModuleId:   old.Id,
			Content:    old.String(),
			Title:      "删除策略",
		}
	})
	b.logX.CreateSysLog(ctx, vobj.ActionDelete, list...)
	return nil
}

// GetStrategyById 获取策略详情
func (b *StrategyBiz) GetStrategyById(ctx context.Context, id uint32) (*bo.StrategyBO, error) {
	wheres := []basescopes.ScopeMethod{
		do.StrategyPreloadEndpoint(),
		do.StrategyPreloadAlarmPages(),
		do.StrategyPreloadCategories(),
		do.StrategyPreloadAlertLevel(),
		do.StrategyPreloadPromNotifies(),
		do.StrategyPreloadPromNotifyUpgrade(),
		do.StrategyPreloadGroupInfo(),
	}
	strategyBO, err := b.strategyRepo.GetStrategyById(ctx, id, wheres...)
	if err != nil {
		return nil, err
	}

	return strategyBO, nil
}

// ListStrategy 获取策略列表
func (b *StrategyBiz) ListStrategy(ctx context.Context, req *bo.ListStrategyRequest) ([]*bo.StrategyBO, error) {
	scopes := []basescopes.ScopeMethod{
		do.StrategyAlertLike(req.Keyword),
		do.StrategyInGroupIds(req.GroupId),
		basescopes.StatusEQ(req.Status),
		do.StrategyPreloadAlertLevel(),
		do.StrategyPreloadCategories(),
		do.StrategyPreloadEndpoint(),
		do.StrategyPreloadGroupInfo(),
		do.StrategyPreloadAlarmPages(),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
		basescopes.InIds(req.StrategyId),
	}

	strategyBOs, err := b.strategyRepo.ListStrategy(ctx, req.Page, scopes...)
	if err != nil {
		return nil, err
	}

	return strategyBOs, nil
}

// SelectStrategy 查询策略
func (b *StrategyBiz) SelectStrategy(ctx context.Context, req *bo.SelectStrategyRequest) ([]*bo.StrategyBO, error) {
	scopes := []basescopes.ScopeMethod{
		do.StrategyAlertLike(req.Keyword),
		basescopes.StatusEQ(req.Status),
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
	}

	strategyBOs, err := b.strategyRepo.ListStrategy(ctx, req.Page, scopes...)
	if err != nil {
		return nil, err
	}

	return strategyBOs, nil
}

// ExportStrategy 导出策略
func (b *StrategyBiz) ExportStrategy(ctx context.Context, req *bo.ExportStrategyRequest) ([]*bo.StrategyBO, error) {
	strategyBOs, err := b.strategyRepo.ListStrategyByIds(ctx, req.Ids)
	if err != nil {
		return nil, err
	}

	list := slices.To(strategyBOs, func(strategyBO *bo.StrategyBO) string {
		return strategyBO.String()
	})
	b.logX.CreateSysLog(ctx, vobj.ActionExport, &bo.SysLogBo{
		ModuleName: vobj.ModuleStrategy,
		ModuleId:   strategyBOs[0].Id,
		Content:    fmt.Sprintf(`{"list":[%s]}`, strings.Join(list, ",")),
		Title:      "导出策略",
	})

	return strategyBOs, nil
}

// GetStrategyWithNotifyObjectById 获取策略详情（包含通知对象）
func (b *StrategyBiz) GetStrategyWithNotifyObjectById(ctx context.Context, id uint32) (*bo.StrategyBO, error) {
	wheres := []basescopes.ScopeMethod{
		do.StrategyPreloadPromNotifies(
			do.PromAlarmNotifyPreloadFieldChatGroups,
			fmt.Sprintf("%s.%s", do.PromAlarmNotifyPreloadFieldBeNotifyMembers, do.PromNotifyMemberPreloadFieldMember),
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
			ModuleName: vobj.ModuleStrategy,
			ModuleId:   strategyBO.Id,
			Content:    fmt.Sprintf(`"notifies":[%s]`, strings.Join(notifyStr, ",")),
			Title:      "绑定策略通知对象",
		}
	})
	b.logX.CreateSysLog(ctx, vobj.ActionUpdate, list...)
	return nil
}
