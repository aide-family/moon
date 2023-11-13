package prombiz

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/api"
	pb "prometheus-manager/api/prom/strategy"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/pkg/model/strategy"
	"prometheus-manager/pkg/util/slices"
)

type (
	StrategyXBiz struct {
		log *log.Helper

		strategyRepo StrategyRepo
	}

	StrategyRepo interface {
		// CreateStrategy 创建策略
		CreateStrategy(ctx context.Context, strategy *biz.StrategyDO) (*biz.StrategyDO, error)
		// UpdateStrategyById 通过id更新策略
		UpdateStrategyById(ctx context.Context, id uint, strategy *biz.StrategyDO) (*biz.StrategyDO, error)
		// BatchUpdateStrategyStatusByIds 通过id批量更新策略状态
		BatchUpdateStrategyStatusByIds(ctx context.Context, status int32, ids []uint) error
		// DeleteStrategyByIds 通过id删除策略
		DeleteStrategyByIds(ctx context.Context, id ...uint) error
		// GetStrategyById 通过id获取策略详情
		GetStrategyById(ctx context.Context, id uint) (*biz.StrategyDO, error)
		// ListStrategy 获取策略列表
		ListStrategy(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*biz.StrategyDO, error)
		// ListStrategyByIds 通过id列表获取策略列表
		ListStrategyByIds(ctx context.Context, ids []uint) ([]*biz.StrategyBO, error)
	}
)

// NewStrategyBiz 创建策略业务对象
func NewStrategyBiz(strategyRepo StrategyRepo, logger log.Logger) *StrategyXBiz {
	return &StrategyXBiz{
		log: log.NewHelper(log.With(logger, "module", "strategy")),

		strategyRepo: strategyRepo,
	}
}

// CreateStrategy 创建策略
func (b *StrategyXBiz) CreateStrategy(ctx context.Context, strategyBO *biz.StrategyBO) (*biz.StrategyBO, error) {
	strategyDO := biz.NewStrategyBO(strategyBO).DO().First()
	strategyDO, err := b.strategyRepo.CreateStrategy(ctx, strategyDO)
	if err != nil {
		return nil, err
	}

	return biz.NewStrategyDO(strategyDO).BO().First(), nil
}

// UpdateStrategyById 更新策略
func (b *StrategyXBiz) UpdateStrategyById(ctx context.Context, id uint32, strategyBO *biz.StrategyBO) (*biz.StrategyBO, error) {
	strategyDO := biz.NewStrategyBO(strategyBO).DO().First()
	strategyDO, err := b.strategyRepo.UpdateStrategyById(ctx, uint(id), strategyDO)
	if err != nil {
		return nil, err
	}

	return biz.NewStrategyDO(strategyDO).BO().First(), nil
}

// BatchUpdateStrategyStatusByIds 批量更新策略状态
func (b *StrategyXBiz) BatchUpdateStrategyStatusByIds(ctx context.Context, status api.Status, ids []uint32) error {
	strategyIds := slices.To(ids, func(t uint32) uint {
		return uint(t)
	})
	return b.strategyRepo.BatchUpdateStrategyStatusByIds(ctx, int32(status), strategyIds)
}

// DeleteStrategyByIds 删除策略
func (b *StrategyXBiz) DeleteStrategyByIds(ctx context.Context, id ...uint32) error {
	if len(id) == 0 {
		return nil
	}
	strategyIds := slices.To(id, func(t uint32) uint {
		return uint(t)
	})

	return b.strategyRepo.DeleteStrategyByIds(ctx, strategyIds...)
}

// GetStrategyById 获取策略详情
func (b *StrategyXBiz) GetStrategyById(ctx context.Context, id uint32) (*biz.StrategyBO, error) {
	strategyDO, err := b.strategyRepo.GetStrategyById(ctx, uint(id))
	if err != nil {
		return nil, err
	}

	return biz.NewStrategyDO(strategyDO).BO().First(), nil
}

// ListStrategy 获取策略列表
func (b *StrategyXBiz) ListStrategy(ctx context.Context, req *pb.ListStrategyRequest) ([]*biz.StrategyBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))

	scopes := []query.ScopeMethod{
		strategy.LikeStrategy(req.GetKeyword()),
		strategy.StatusEQ(int32(req.GetStatus())),
	}

	strategyDOs, err := b.strategyRepo.ListStrategy(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return biz.NewStrategyDO(strategyDOs...).BO().List(), pgInfo, nil
}

// SelectStrategy 查询策略
func (b *StrategyXBiz) SelectStrategy(ctx context.Context, req *pb.SelectStrategyRequest) ([]*biz.StrategyBO, query.Pagination, error) {
	pgReq := req.GetPage()
	pgInfo := query.NewPage(int(pgReq.GetCurr()), int(pgReq.GetSize()))

	scopes := []query.ScopeMethod{
		strategy.LikeStrategy(req.GetKeyword()),
		strategy.StatusEQ(int32(api.Status_STATUS_ENABLED)),
	}

	strategyDOs, err := b.strategyRepo.ListStrategy(ctx, pgInfo, scopes...)
	if err != nil {
		return nil, nil, err
	}

	return biz.NewStrategyDO(strategyDOs...).BO().List(), pgInfo, nil
}

// ExportStrategy 导出策略
func (b *StrategyXBiz) ExportStrategy(ctx context.Context, req *pb.ExportStrategyRequest) ([]*biz.StrategyBO, error) {
	strategyIds := slices.To(req.GetIds(), func(t uint32) uint {
		return uint(t)
	})

	strategyBOs, err := b.strategyRepo.ListStrategyByIds(ctx, strategyIds)
	if err != nil {
		return nil, err
	}

	return strategyBOs, nil
}
