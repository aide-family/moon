package biz

import (
	"context"

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
	}
)

// NewStrategyBiz 创建策略业务对象
func NewStrategyBiz(strategyRepo repository.StrategyRepo, notifyRepo repository.NotifyRepo, logger log.Logger) *StrategyBiz {
	return &StrategyBiz{
		log: log.NewHelper(log.With(logger, "module", "strategy")),

		strategyRepo: strategyRepo,
		notifyRepo:   notifyRepo,
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

	return strategyBO, nil
}

// UpdateStrategyById 更新策略
func (b *StrategyBiz) UpdateStrategyById(ctx context.Context, id uint32, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	strategyBO, err := b.strategyRepo.UpdateStrategyById(ctx, id, strategyBO)
	if err != nil {
		return nil, err
	}

	return strategyBO, nil
}

// BatchUpdateStrategyStatusByIds 批量更新策略状态
func (b *StrategyBiz) BatchUpdateStrategyStatusByIds(ctx context.Context, status api.Status, ids []uint32) error {
	return b.strategyRepo.BatchUpdateStrategyStatusByIds(ctx, vo.Status(status), ids)
}

// DeleteStrategyByIds 删除策略
func (b *StrategyBiz) DeleteStrategyByIds(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}
	return b.strategyRepo.DeleteStrategyByIds(ctx, ids...)
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
	return b.strategyRepo.BindStrategyNotifyObject(ctx, strategyBO, notifyBOs)
}
