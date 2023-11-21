package promdict

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/dobo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
)

var _ repository.PromDictRepo = (*promDictRepoImpl)(nil)

type promDictRepoImpl struct {
	data *data.Data
	log  *log.Helper

	query.IAction[model.PromDict]
}

func (l *promDictRepoImpl) CreateDict(ctx context.Context, dict *dobo.DictDO) (*dobo.DictDO, error) {
	newModelData := dobo.DictDOToModel(dict)
	if err := l.WithContext(ctx).Create(newModelData); err != nil {
		return nil, err
	}

	return dobo.DictModelToDO(newModelData), nil
}

func (l *promDictRepoImpl) UpdateDictById(ctx context.Context, id uint, dict *dobo.DictDO) (*dobo.DictDO, error) {
	newModelData := dobo.DictDOToModel(dict)
	if err := l.WithContext(ctx).UpdateByID(id, newModelData); err != nil {
		return nil, err
	}

	newModelData, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}

	return dobo.DictModelToDO(newModelData), nil
}

func (l *promDictRepoImpl) BatchUpdateDictStatusByIds(ctx context.Context, status int32, ids []uint) error {
	if err := l.WithContext(ctx).Update(&model.PromDict{Status: status}, query.WhereID(ids...)); err != nil {
		return err
	}
	return nil
}

func (l *promDictRepoImpl) DeleteDictByIds(ctx context.Context, id ...uint) error {
	if err := l.WithContext(ctx).Delete(query.WhereID(id...)); err != nil {
		return err
	}
	return nil
}

func (l *promDictRepoImpl) GetDictById(ctx context.Context, id uint) (*dobo.DictDO, error) {
	detailModel, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return dobo.DictModelToDO(detailModel), nil
}

func (l *promDictRepoImpl) ListDict(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*dobo.DictDO, error) {
	dictModelList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	boList := make([]*dobo.DictDO, 0, len(dictModelList))
	for _, m := range dictModelList {
		boList = append(boList, dobo.DictModelToDO(m))
	}

	return boList, nil
}

func NewPromDictRepo(data *data.Data, logger log.Logger) repository.PromDictRepo {
	return &promDictRepoImpl{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data.repository.dict")),
		IAction: query.NewAction[model.PromDict](
			query.WithDB[model.PromDict](data.DB()),
		),
	}
}
