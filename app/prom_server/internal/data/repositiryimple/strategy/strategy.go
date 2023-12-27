package strategy

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/basescopes"
	"prometheus-manager/pkg/helper/valueobj"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
)

var _ repository.StrategyRepo = (*strategyRepoImpl)(nil)

type (
	strategyRepoImpl struct {
		repository.UnimplementedStrategyRepo
		query.IAction[model.PromStrategy]

		data *data.Data
		log  *log.Helper

		strategyGroupRepo repository.StrategyGroupRepo
	}
)

func (l *strategyRepoImpl) ListStrategyByIds(ctx context.Context, ids []uint32) ([]*bo.StrategyBO, error) {
	modelList := make([]*model.PromStrategy, 0, len(ids))
	if err := l.WithContext(ctx).DB().Find(&modelList).Error; err != nil {
		return nil, err
	}

	list := make([]*bo.StrategyBO, 0, len(modelList))
	for _, m := range modelList {
		list = append(list, bo.StrategyModelToBO(m))
	}
	return list, nil
}

func (l *strategyRepoImpl) CreateStrategy(ctx context.Context, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	newStrategy := strategyBO.ToModel()
	if err := l.WithContext(ctx).Create(newStrategy); err != nil {
		return nil, err
	}

	// 更新策略组的策略数量
	if err := l.strategyGroupRepo.UpdateStrategyCount(ctx, strategyBO.GroupId); err != nil {
		return nil, err
	}

	// 更新策略组的启用策略数量
	if err := l.strategyGroupRepo.UpdateEnableStrategyCount(ctx, strategyBO.GroupId); err != nil {
		return nil, err
	}

	return bo.StrategyModelToBO(newStrategy), nil
}

func (l *strategyRepoImpl) UpdateStrategyById(ctx context.Context, id uint32, strategyBO *bo.StrategyBO) (*bo.StrategyBO, error) {
	detail, err := l.GetStrategyById(ctx, id)
	if err != nil {
		return nil, err
	}

	newStrategy := strategyBO.ToModel()
	if err = l.WithContext(ctx).UpdateByID(id, newStrategy); err != nil {
		return nil, err
	}

	if detail.Status != newStrategy.Status && !newStrategy.Status.IsUnknown() {
		// 更新策略组的启用策略数量
		if err = l.strategyGroupRepo.UpdateEnableStrategyCount(ctx, strategyBO.GroupId); err != nil {
			return nil, err
		}
	}

	return bo.StrategyModelToBO(newStrategy), nil
}

func (l *strategyRepoImpl) BatchUpdateStrategyStatusByIds(ctx context.Context, status valueobj.Status, ids []uint32) error {
	if err := l.WithContext(ctx).Update(&model.PromStrategy{Status: status}, basescopes.InIds(ids...)); err != nil {
		return err
	}
	// 更新策略组的启用策略数量
	if err := l.strategyGroupRepo.UpdateEnableStrategyCount(ctx, ids...); err != nil {
		return err
	}
	return nil
}

func (l *strategyRepoImpl) DeleteStrategyByIds(ctx context.Context, ids ...uint32) error {
	var detailList []*model.PromStrategy
	if err := l.data.DB().Scopes(basescopes.InIds(ids...)).Find(&detailList).Error; err != nil {
		return err
	}

	if err := l.WithContext(ctx).Delete(basescopes.InIds(ids...)); err != nil {
		return err
	}
	groupIds := slices.To(detailList, func(i *model.PromStrategy) uint32 {
		return i.GroupID
	})
	// 更新策略组的策略数量
	if err := l.strategyGroupRepo.UpdateStrategyCount(ctx, groupIds...); err != nil {
		return err
	}

	// 更新策略组的启用策略数量
	if err := l.strategyGroupRepo.UpdateEnableStrategyCount(ctx, groupIds...); err != nil {
		return err
	}
	return nil
}

func (l *strategyRepoImpl) GetStrategyById(ctx context.Context, id uint32, wheres ...query.ScopeMethod) (*bo.StrategyBO, error) {
	firstStrategy, err := l.WithContext(ctx).FirstByID(id, wheres...)
	if err != nil {
		return nil, err
	}
	return bo.StrategyModelToBO(firstStrategy), nil
}

func (l *strategyRepoImpl) ListStrategy(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.StrategyBO, error) {
	listStrategy, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(listStrategy, func(i *model.PromStrategy) *bo.StrategyBO {
		if i == nil {
			return nil
		}
		return bo.StrategyModelToBO(i)
	})
	return list, nil
}

func NewStrategyRepo(data *data.Data, strategyGroupRepo repository.StrategyGroupRepo, logger log.Logger) repository.StrategyRepo {
	return &strategyRepoImpl{
		data: data,
		log:  log.NewHelper(logger),
		IAction: query.NewAction[model.PromStrategy](
			query.WithDB[model.PromStrategy](data.DB()),
		),
		strategyGroupRepo: strategyGroupRepo,
	}
}
