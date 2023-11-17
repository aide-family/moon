package strategy

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"

	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	data2 "prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/model"
)

var _ repository.StrategyRepo = (*strategyRepoImpl)(nil)

type (
	strategyRepoImpl struct {
		query.IAction[model.PromStrategy]

		data *data2.Data
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

func (l *strategyRepoImpl) CreateStrategy(ctx context.Context, strategy *dobo.StrategyDO) (*dobo.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *strategyRepoImpl) UpdateStrategyById(ctx context.Context, id uint, strategy *dobo.StrategyDO) (*dobo.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *strategyRepoImpl) BatchUpdateStrategyStatusByIds(ctx context.Context, status int32, ids []uint) error {
	//TODO implement me
	panic("implement me")
}

func (l *strategyRepoImpl) DeleteStrategyByIds(ctx context.Context, id ...uint) error {
	//TODO implement me
	panic("implement me")
}

func (l *strategyRepoImpl) GetStrategyById(ctx context.Context, id uint) (*dobo.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *strategyRepoImpl) ListStrategy(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func NewStrategyRepo(data *data2.Data, logger log.Logger) repository.StrategyRepo {
	return &strategyRepoImpl{
		data: data,
		log:  log.NewHelper(logger),
		IAction: query.NewAction[model.PromStrategy](
			query.WithDB[model.PromStrategy](data.DB()),
		),
	}
}
