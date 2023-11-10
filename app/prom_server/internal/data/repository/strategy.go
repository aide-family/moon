package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/prombiz"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/model"
)

type (
	strategyRepoImpl struct {
		query.IAction[model.PromStrategy]

		data *data.Data
		log  *log.Helper
	}
)

func (l *strategyRepoImpl) CreateStrategy(ctx context.Context, strategy *biz.StrategyDO) (*biz.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *strategyRepoImpl) UpdateStrategyById(ctx context.Context, id uint, strategy *biz.StrategyDO) (*biz.StrategyDO, error) {
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

func (l *strategyRepoImpl) GetStrategyById(ctx context.Context, id uint) (*biz.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func (l *strategyRepoImpl) ListStrategy(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*biz.StrategyDO, error) {
	//TODO implement me
	panic("implement me")
}

func NewStrategyRepo(data *data.Data, logger log.Logger) prombiz.StrategyRepo {
	return &strategyRepoImpl{
		data: data,
		log:  log.NewHelper(logger),
	}
}
