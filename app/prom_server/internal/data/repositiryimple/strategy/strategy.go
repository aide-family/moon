package strategy

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/strategy"
	"prometheus-manager/pkg/util/slices"

	"prometheus-manager/app/prom_server/internal/biz/dobo"
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
	}
)

func (l *strategyRepoImpl) ListStrategyByIds(ctx context.Context, ids []uint) ([]*dobo.StrategyDO, error) {
	modelList := make([]*model.PromStrategy, 0, len(ids))
	if err := l.WithContext(ctx).DB().Find(&modelList).Error; err != nil {
		return nil, err
	}

	list := make([]*dobo.StrategyDO, 0, len(modelList))
	for _, m := range modelList {
		list = append(list, dobo.StrategyModelToDO(m))
	}
	return list, nil
}

func (l *strategyRepoImpl) CreateStrategy(ctx context.Context, strategyDO *dobo.StrategyDO) (*dobo.StrategyDO, error) {
	newStrategy := strategyDO.ToModel()
	if err := l.WithContext(ctx).Create(newStrategy); err != nil {
		return nil, err
	}
	return dobo.StrategyModelToDO(newStrategy), nil
}

func (l *strategyRepoImpl) UpdateStrategyById(ctx context.Context, id uint, strategyDO *dobo.StrategyDO) (*dobo.StrategyDO, error) {
	newStrategy := strategyDO.ToModel()
	if err := l.WithContext(ctx).UpdateByID(id, newStrategy); err != nil {
		return nil, err
	}
	return dobo.StrategyModelToDO(newStrategy), nil
}

func (l *strategyRepoImpl) BatchUpdateStrategyStatusByIds(ctx context.Context, status int32, ids []uint) error {
	if err := l.WithContext(ctx).Update(&model.PromStrategy{Status: status}, strategy.InIds(ids)); err != nil {
		return err
	}
	return nil
}

func (l *strategyRepoImpl) DeleteStrategyByIds(ctx context.Context, id ...uint) error {
	if err := l.WithContext(ctx).Delete(strategy.InIds(id)); err != nil {
		return err
	}
	return nil
}

func (l *strategyRepoImpl) GetStrategyById(ctx context.Context, id uint) (*dobo.StrategyDO, error) {
	firstStrategy, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return dobo.StrategyModelToDO(firstStrategy), nil
}

func (l *strategyRepoImpl) ListStrategy(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.StrategyDO, error) {
	listStrategy, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(listStrategy, func(i *model.PromStrategy) *dobo.StrategyDO {
		if i == nil {
			return nil
		}
		return dobo.StrategyModelToDO(i)
	})
	return list, nil
}

func NewStrategyRepo(data *data.Data, logger log.Logger) repository.StrategyRepo {
	return &strategyRepoImpl{
		data: data,
		log:  log.NewHelper(logger),
		IAction: query.NewAction[model.PromStrategy](
			query.WithDB[model.PromStrategy](data.DB()),
		),
	}
}
