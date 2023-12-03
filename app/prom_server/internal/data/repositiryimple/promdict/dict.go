package promdict

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"prometheus-manager/app/prom_server/internal/biz/bo"
	"prometheus-manager/app/prom_server/internal/biz/repository"
	"prometheus-manager/app/prom_server/internal/data"
	"prometheus-manager/pkg/helper/model"
	"prometheus-manager/pkg/helper/valueobj"
)

var _ repository.PromDictRepo = (*promDictRepoImpl)(nil)

type promDictRepoImpl struct {
	repository.UnimplementedPromDictRepo
	data *data.Data
	log  *log.Helper

	query.IAction[model.PromDict]
}

func (l *promDictRepoImpl) CreateDict(ctx context.Context, dictBO *bo.DictBO) (*bo.DictBO, error) {
	newModelData := dictBO.ToModel()
	if err := l.WithContext(ctx).Create(newModelData); err != nil {
		return nil, err
	}

	return bo.DictModelToBO(newModelData), nil
}

func (l *promDictRepoImpl) UpdateDictById(ctx context.Context, id uint, dictBO *bo.DictBO) (*bo.DictBO, error) {
	newModelData := dictBO.ToModel()
	if err := l.WithContext(ctx).UpdateByID(id, newModelData); err != nil {
		return nil, err
	}

	newModelData, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}

	return bo.DictModelToBO(newModelData), nil
}

func (l *promDictRepoImpl) BatchUpdateDictStatusByIds(ctx context.Context, status valueobj.Status, ids []uint) error {
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

func (l *promDictRepoImpl) GetDictById(ctx context.Context, id uint) (*bo.DictBO, error) {
	detailModel, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return bo.DictModelToBO(detailModel), nil
}

func (l *promDictRepoImpl) ListDict(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*bo.DictBO, error) {
	dictModelList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	boList := make([]*bo.DictBO, 0, len(dictModelList))
	for _, m := range dictModelList {
		boList = append(boList, bo.DictModelToBO(m))
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
