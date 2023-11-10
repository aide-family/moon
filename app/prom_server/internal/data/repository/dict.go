package repository

import (
	"context"

	query "github.com/aide-cloud/gorm-normalize"
	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/plugin/soft_delete"
	"prometheus-manager/pkg/model"

	"prometheus-manager/app/prom_server/internal/biz"
	"prometheus-manager/app/prom_server/internal/biz/dictbiz"
	"prometheus-manager/app/prom_server/internal/data"
)

type dictRepoImpl struct {
	data *data.Data
	log  *log.Helper

	query.IAction[model.PromDict]
}

func (l *dictRepoImpl) CreateDict(ctx context.Context, dict *biz.DictDO) (*biz.DictDO, error) {
	newModelData := dictDOToModel(dict)
	if err := l.WithContext(ctx).Create(newModelData); err != nil {
		return nil, err
	}

	return dictModelToDO(newModelData), nil
}

func (l *dictRepoImpl) UpdateDictById(ctx context.Context, id uint, dict *biz.DictDO) (*biz.DictDO, error) {
	newModelData := dictDOToModel(dict)
	if err := l.WithContext(ctx).UpdateByID(id, newModelData); err != nil {
		return nil, err
	}

	newModelData, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}

	return dictModelToDO(newModelData), nil
}

func (l *dictRepoImpl) BatchUpdateDictStatusByIds(ctx context.Context, status int32, ids []uint) error {
	if err := l.WithContext(ctx).Update(&model.PromDict{Status: status}, query.WhereID(ids...)); err != nil {
		return err
	}
	return nil
}

func (l *dictRepoImpl) DeleteDictByIds(ctx context.Context, id ...uint) error {
	if err := l.WithContext(ctx).Delete(query.WhereID(id...)); err != nil {
		return err
	}
	return nil
}

func (l *dictRepoImpl) GetDictById(ctx context.Context, id uint) (*biz.DictDO, error) {
	detailModel, err := l.WithContext(ctx).FirstByID(id)
	if err != nil {
		return nil, err
	}
	return dictModelToDO(detailModel), nil
}

func (l *dictRepoImpl) ListDict(ctx context.Context, pgInfo query.Pagination, scopes ...query.ScopeMethod) ([]*biz.DictDO, error) {
	dictModelList, err := l.WithContext(ctx).List(pgInfo, scopes...)
	if err != nil {
		return nil, err
	}

	boList := make([]*biz.DictDO, 0, len(dictModelList))
	for _, m := range dictModelList {
		boList = append(boList, dictModelToDO(m))
	}

	return boList, nil
}

// dictModelToDO dict model to dict do
func dictModelToDO(m *model.PromDict) *biz.DictDO {
	if m == nil {
		return nil
	}
	return &biz.DictDO{
		Id:        m.ID,
		Name:      m.Name,
		Category:  m.Category,
		Status:    m.Status,
		Remark:    m.Remark,
		Color:     m.Color,
		CreatedAt: m.CreatedAt,
		UpdatedAt: m.UpdatedAt,
		DeletedAt: int64(m.DeletedAt),
	}
}

// dictDOToModel dict do to dict model
func dictDOToModel(d *biz.DictDO) *model.PromDict {
	if d == nil {
		return nil
	}
	return &model.PromDict{
		Name:     d.Name,
		Category: d.Category,
		Status:   d.Status,
		Remark:   d.Remark,
		Color:    d.Color,
		BaseModel: query.BaseModel{
			ID:        d.Id,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
			DeletedAt: soft_delete.DeletedAt(d.DeletedAt),
		},
	}
}

func NewDictRepo(data *data.Data, logger log.Logger) dictbiz.Repo {
	return &dictRepoImpl{
		data: data,
		log:  log.NewHelper(log.With(logger, "module", "data.repository.dict")),
		IAction: query.NewAction[model.PromDict](
			query.WithDB[model.PromDict](data.DB()),
		),
	}
}
