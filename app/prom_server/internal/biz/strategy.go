package biz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api"
	strategyPB "prometheus-manager/api/prom/strategy"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/do/basescopes"
	"prometheus-manager/app/prom_server/internal/biz/do/strategyscopes"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/biz/vo"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyXBiz struct {
		log *log.Helper

		strategyRepo repository.StrategyRepo
	}
)

// NewStrategyBiz 创建策略业务对象
func NewStrategyBiz(strategyRepo repository.StrategyRepo, logger log.Logger) *StrategyXBiz {
	return &StrategyXBiz{
		log: log.NewHelper(log.With(logger, "module", "strategy")),

		strategyRepo: strategyRepo,
	}
}

// CreateStrategy 创建策略
func (b *StrategyXBiz) CreateStrategy(ctx context.Context, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
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
func (b *StrategyXBiz) UpdateStrategyById(ctx context.Context, id uint32, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	strategyBO, err := b.strategyRepo.UpdateStrategyById(ctx, id, strategyBO)
	if err != nil {
		return nil, err
	}

	return strategyBO, nil
}

// BatchUpdateStrategyStatusByIds 批量更新策略状态
func (b *StrategyXBiz) BatchUpdateStrategyStatusByIds(ctx context.Context, status api.Status, ids []uint32) error {
	return b.strategyRepo.BatchUpdateStrategyStatusByIds(ctx, vo.Status(status), ids)
}

// DeleteStrategyByIds 删除策略
func (b *StrategyXBiz) DeleteStrategyByIds(ctx context.Context, ids ...uint32) error {
	if len(ids) == 0 {
		return nil
	}
	return b.strategyRepo.DeleteStrategyByIds(ctx, ids...)
}

// GetStrategyById 获取策略详情
func (b *StrategyXBiz) GetStrategyById(ctx context.Context, id uint32) (*bo.StrategyBO, error) {
	wheres := []query.ScopeMethod{
		strategyscopes.PreloadEndpoint,
		strategyscopes.PreloadAlarmPages,
		strategyscopes.PreloadCategories,
		strategyscopes.PreloadAlertLevel,
		strategyscopes.PreloadPromNotifies,
		strategyscopes.PreloadPromNotifyUpgrade,
		strategyscopes.PreloadGroupInfo,
	}
	strategyBO, err := b.strategyRepo.GetStrategyById(ctx, id, wheres...)
	if err != nil {
		return nil, err
	}

	return strategyBO, nil
}

// ListStrategy 获取策略列表
func (b *StrategyXBiz) ListStrategy(ctx context.Context, req *strategyPB.ListStrategyRequest) ([]*bo.StrategyBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	scopes := []query.ScopeMethod{
		strategyscopes.AlertLike(req.GetKeyword()),
		strategyscopes.GroupIdsEQ(req.GetGroupId()),
		basescopes.StatusEQ(vo.Status(req.GetStatus())),
		strategyscopes.PreloadAlertLevel,
		strategyscopes.PreloadCategories,
		strategyscopes.PreloadEndpoint,
		basescopes.UpdateAtDesc(),
		basescopes.CreatedAtDesc(),
	}

	strategyBOs, err := b.strategyRepo.ListStrategy(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return strategyBOs, pgInfo, nil
}

// SelectStrategy 查询策略
func (b *StrategyXBiz) SelectStrategy(ctx context.Context, req *strategyPB.SelectStrategyRequest) ([]*bo.StrategyBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(pgReq.GetCurr(), pgReq.GetSize())

	scopes := []query.ScopeMethod{
		strategyscopes.AlertLike(req.GetKeyword()),
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
func (b *StrategyXBiz) ExportStrategy(ctx context.Context, req *strategyPB.ExportStrategyRequest) ([]*bo.StrategyBO, error) {
	strategyBOs, err := b.strategyRepo.ListStrategyByIds(ctx, req.GetIds())
	if err != nil {
		return nil, err
	}

	return strategyBOs, nil
}
