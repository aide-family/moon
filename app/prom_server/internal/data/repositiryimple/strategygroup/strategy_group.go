package strategygroup

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/model/strategygroup"
	"prometheus-manager/pkg/util/slices"
)

var _ repository.StrategyGroupRepo = (*strategyGroupRepoImpl)(nil)

type (
	strategyGroupRepoImpl struct {
		repository.UnimplementedStrategyGroupRepo
		query.IAction[model.PromStrategyGroup]

		data *data.Data
		log  *log.Helper
	}
)

func (l *strategyGroupRepoImpl) Create(ctx context.Context, strategyGroup *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error) {
	strategyGroupModel := strategyGroup.ToModel()
	if err := l.WithContext(ctx).Create(strategyGroupModel); err != nil {
		return nil, err
	}
	return dobo.StrategyGroupModelToDO(strategyGroupModel), nil
}

func (l *strategyGroupRepoImpl) UpdateById(ctx context.Context, id uint, strategyGroup *dobo.StrategyGroupDO) (*dobo.StrategyGroupDO, error) {
	strategyGroupModel := strategyGroup.ToModel()
	if err := l.WithContext(ctx).UpdateByID(id, strategyGroupModel); err != nil {
		return nil, err
	}
	return dobo.StrategyGroupModelToDO(strategyGroupModel), nil
}

func (l *strategyGroupRepoImpl) BatchUpdateStatus(ctx context.Context, status int32, ids []uint) error {
	if err := l.WithContext(ctx).Update(&model.PromStrategyGroup{Status: status}, strategygroup.InIds(ids)); err != nil {
		return err
	}
	return nil
}

func (l *strategyGroupRepoImpl) DeleteByIds(ctx context.Context, ids ...uint) error {
	if err := l.WithContext(ctx).Delete(strategygroup.InIds(ids)); err != nil {
		return err
	}
	return nil
}

func (l *strategyGroupRepoImpl) GetById(ctx context.Context, id uint) (*dobo.StrategyGroupDO, error) {
	first, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return dobo.StrategyGroupModelToDO(first), nil
}

func (l *strategyGroupRepoImpl) List(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.StrategyGroupDO, error) {
	strategyModelList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}
	list := slices.To(strategyModelList, func(m *model.PromStrategyGroup) *dobo.StrategyGroupDO {
		return dobo.StrategyGroupModelToDO(m)
	})
	return list, nil
}

func NewStrategyGroupRepo(data *data.Data, logger log.Logger) repository.StrategyGroupRepo {
	return &strategyGroupRepoImpl{
		IAction: query.NewAction[model.PromStrategyGroup](
			query.WithDB[model.PromStrategyGroup](data.DB()),
		),
		data: data,
		log:  log.NewHelper(logger),
	}
}
