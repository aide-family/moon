package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model"
	"github.com/aide-family/moon/pkg/palace/model/query"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"gorm.io/gen"
)

func NewDictRepository(data *data.Data) repository.Dict {
	return &dictRepositoryImpl{
		data: data,
	}
}

type dictRepositoryImpl struct {
	data *data.Data
}

func (l *dictRepositoryImpl) UpdateStatusByIds(ctx context.Context, status vobj.Status, ids ...uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysDict.Where(query.SysDict.ID.In(ids...)).Update(query.SysDict.Status, status)
	return err
}

func (l *dictRepositoryImpl) DeleteByID(ctx context.Context, id uint32) error {
	_, err := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysDict.Where(query.SysDict.ID.Eq(id)).Delete()
	return err
}

func (l *dictRepositoryImpl) Create(ctx context.Context, dict *bo.CreateDictParams) (*model.SysDict, error) {
	dictModel := createDictParamsToModel(dict)
	dictModel.WithContext(ctx)
	q := query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysDict
	if err := q.Create(dictModel); !types.IsNil(err) {
		return nil, err
	}
	return dictModel, nil
}

func (l *dictRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryDictListParams) ([]*model.SysDict, error) {
	queryWrapper := query.Use(l.data.GetMainDB(ctx)).SysDict.WithContext(ctx)

	var wheres []gen.Condition
	if !params.Status.IsUnknown() {
		wheres = append(wheres, query.SysDict.Status.Eq(params.Status.GetValue()))
	}

	if !params.DictType.IsUnknown() {
		wheres = append(wheres, query.SysDict.DictType.Eq(params.DictType.GetValue()))
	}

	if !types.TextIsNull(params.Keyword) {
		queryWrapper = queryWrapper.Or(
			query.SysDict.Name.Like(params.Keyword),
			query.SysDict.Value.Like(params.Keyword),
			query.SysDict.Remark.Like(params.Keyword),
		)
	}
	queryWrapper = queryWrapper.Where(wheres...)
	if err := types.WithPageQuery[query.ISysDictDo](queryWrapper, params.Page); err != nil {
		return nil, err
	}
	return queryWrapper.Order(query.SysDict.ID.Desc()).Find()
}

func (l *dictRepositoryImpl) BatchCreate(ctx context.Context, createDicts []*bo.CreateDictParams) error {

	dictModels := types.SliceToWithFilter(createDicts, func(item *bo.CreateDictParams) (*model.SysDict, bool) {
		if types.IsNil(item) || types.TextIsNull(item.Name) {
			return nil, false
		}
		return createDictParamsToModel(item), true
	})
	return query.Use(l.data.GetMainDB(ctx)).WithContext(ctx).SysDict.CreateInBatches(dictModels, 10)
}

func (l *dictRepositoryImpl) GetByID(ctx context.Context, id uint32) (*model.SysDict, error) {
	return query.Use(l.data.GetMainDB(ctx)).SysDict.WithContext(ctx).Where(query.SysDict.ID.Eq(id)).First()
}

func (l *dictRepositoryImpl) UpdateByID(ctx context.Context, dict *bo.UpdateDictParams) error {
	updateParam := dict.UpdateParam
	_, err := query.Use(l.data.GetMainDB(ctx)).SysDict.WithContext(ctx).Where(query.SysDict.ID.Eq(dict.ID)).UpdateSimple(
		query.SysDict.Name.Value(updateParam.Name),
		query.SysDict.Value.Value(updateParam.Value),
		query.SysDict.CssClass.Value(updateParam.CssClass),
		query.SysDict.ColorType.Value(updateParam.ColorType),
		query.SysDict.Remark.Value(updateParam.Remark),
		query.SysDict.ImageUrl.Value(updateParam.ImageUrl),
		query.SysDict.Icon.Value(updateParam.Icon),
	)
	return err

}

// createDictParamsToModel create dict params to model
func createDictParamsToModel(dict *bo.CreateDictParams) *model.SysDict {
	if types.IsNil(dict) {
		return nil
	}
	return &model.SysDict{
		Name:         dict.Name,
		Value:        dict.Value,
		DictType:     dict.DictType,
		ColorType:    dict.ColorType,
		CssClass:     dict.CssClass,
		Status:       dict.Status,
		Remark:       dict.Remark,
		Icon:         dict.Icon,
		ImageUrl:     dict.ImageUrl,
		LanguageCode: dict.LanguageCode,
	}
}
