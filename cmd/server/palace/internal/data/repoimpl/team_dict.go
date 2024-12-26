package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/imodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"gorm.io/gen/field"

	"gorm.io/gen"
)

// NewTeamDictRepository 创建数据库字典操作
func NewTeamDictRepository(data *data.Data, cacheRepo repository.Cache) repository.TeamDict {
	return &teamDictRepositoryImpl{
		data:      data,
		cacheRepo: cacheRepo,
	}
}

type teamDictRepositoryImpl struct {
	data      *data.Data
	cacheRepo repository.Cache
}

func (l *teamDictRepositoryImpl) cacheDict(ctx context.Context, id uint32) {
	if id <= 0 {
		return
	}
	iDict, err := l.GetByID(ctx, id)
	if err != nil {
		return
	}
	l.cacheRepo.AppendDict(ctx, iDict, true)
}

func (l *teamDictRepositoryImpl) cacheDictItemList(ctx context.Context, items []imodel.IDict) {
	if len(items) == 0 {
		return
	}

	l.cacheRepo.AppendDictList(ctx, items, true)
}

func (l *teamDictRepositoryImpl) cacheDictList(ctx context.Context, ids []uint32) {
	if len(ids) == 0 {
		return
	}
	iDictList, err := l.GetByIDs(ctx, ids)
	if err != nil {
		return
	}
	l.cacheRepo.AppendDictList(ctx, iDictList, true)
}

func (l *teamDictRepositoryImpl) UpdateStatusByIds(ctx context.Context, params *bo.UpdateDictStatusParams) error {
	ids := params.IDs
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	defer l.cacheDictList(ctx, params.IDs)
	_, err = bizQuery.SysDict.WithContext(ctx).Where(bizQuery.SysDict.ID.In(ids...)).Update(bizQuery.SysDict.Status, params.Status)
	return err
}

func (l *teamDictRepositoryImpl) DeleteByID(ctx context.Context, id uint32) error {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	defer l.cacheDict(ctx, id)
	_, err = bizQuery.SysDict.Where(bizQuery.SysDict.ID.Eq(id)).Delete()
	return err
}

func (l *teamDictRepositoryImpl) Create(ctx context.Context, dict *bo.CreateDictParams) (imodel.IDict, error) {
	dictDo, err := l.createBizDictModel(ctx, dict)
	if !types.IsNil(err) {
		return nil, err
	}
	defer l.cacheDict(ctx, dictDo.GetID())
	return dictDo, err
}

func (l *teamDictRepositoryImpl) FindByPage(ctx context.Context, params *bo.QueryDictListParams) ([]imodel.IDict, error) {
	return l.listBizDictModel(ctx, params)
}

func (l *teamDictRepositoryImpl) GetByID(ctx context.Context, id uint32) (imodel.IDict, error) {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.SysDict.WithContext(ctx)
	return bizWrapper.Where(bizQuery.SysDict.ID.Eq(id)).Preload(field.Associations).First()
}

func (l *teamDictRepositoryImpl) GetByIDs(ctx context.Context, ids []uint32) ([]imodel.IDict, error) {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.SysDict.WithContext(ctx)
	sysDictList, err := bizWrapper.Where(bizQuery.SysDict.ID.In(ids...)).Preload(field.Associations).Find()
	if err != nil {
		return nil, err
	}

	return types.SliceTo(sysDictList, func(item *bizmodel.SysDict) imodel.IDict {
		return imodel.IDict(item)
	}), nil
}

func (l *teamDictRepositoryImpl) UpdateByID(ctx context.Context, dict *bo.UpdateDictParams) error {
	defer l.cacheDict(ctx, dict.ID)
	return l.updateBizDictModel(ctx, dict)
}

func (l *teamDictRepositoryImpl) listBizDictModel(ctx context.Context, params *bo.QueryDictListParams) ([]imodel.IDict, error) {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	bizWrapper := bizQuery.SysDict.WithContext(ctx)

	var wheres []gen.Condition

	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.SysDict.Status.Eq(params.Status.GetValue()))
	}

	if !params.DictType.IsUnknown() {
		wheres = append(wheres, bizQuery.SysDict.DictType.Eq(params.DictType.GetValue()))
	}

	if !types.TextIsNull(params.Keyword) {
		bizWrapper = bizWrapper.Or(bizQuery.SysDict.Name.Like(params.Keyword))
		bizWrapper = bizWrapper.Or(bizQuery.SysDict.Value.Like(params.Keyword))
		bizWrapper = bizWrapper.Or(bizQuery.SysDict.Remark.Like(params.Keyword))
	}

	bizWrapper = bizWrapper.Where(wheres...).Preload(field.Associations)

	if bizWrapper, err = types.WithPageQuery(bizWrapper, params.Page); err != nil {
		return nil, err
	}
	sysDictList, err := bizWrapper.Order(bizQuery.SysDict.ID.Desc()).Find()
	if !types.IsNil(err) {
		return nil, err
	}
	dictList := types.SliceTo(sysDictList, func(dict *bizmodel.SysDict) imodel.IDict {
		return dict
	})
	l.cacheDictItemList(ctx, dictList)
	return dictList, nil
}

// createBizDictModel create team dict model
func (l *teamDictRepositoryImpl) createBizDictModel(ctx context.Context, dict *bo.CreateDictParams) (*bizmodel.SysDict, error) {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	dictBizModel := createBizDictParamsToModel(ctx, dict)
	if err := bizQuery.SysDict.WithContext(ctx).Create(dictBizModel); !types.IsNil(err) {
		return nil, err
	}
	return dictBizModel, nil
}

func (l *teamDictRepositoryImpl) updateBizDictModel(ctx context.Context, params *bo.UpdateDictParams) error {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	updateParam := params.UpdateParam
	id := params.ID
	_, err = bizQuery.SysDict.Where(bizQuery.SysDict.ID.Eq(id)).UpdateSimple(
		bizQuery.SysDict.Name.Value(updateParam.Name),
		bizQuery.SysDict.Remark.Value(updateParam.Remark),
		bizQuery.SysDict.Value.Value(updateParam.Value),
		bizQuery.SysDict.CSSClass.Value(updateParam.CSSClass),
		bizQuery.SysDict.ColorType.Value(updateParam.ColorType),
		bizQuery.SysDict.ImageURL.Value(updateParam.ImageURL),
		bizQuery.SysDict.Icon.Value(updateParam.Icon),
	)
	return err
}

func createBizDictParamsToModel(ctx context.Context, dict *bo.CreateDictParams) *bizmodel.SysDict {
	if types.IsNil(dict) {
		return nil
	}
	modelDict := &bizmodel.SysDict{
		Name:         dict.Name,
		Value:        dict.Value,
		DictType:     dict.DictType,
		ColorType:    dict.ColorType,
		CSSClass:     dict.CSSClass,
		Status:       dict.Status,
		Remark:       dict.Remark,
		Icon:         dict.Icon,
		ImageURL:     dict.ImageURL,
		LanguageCode: dict.LanguageCode,
	}
	modelDict.WithContext(ctx)
	return modelDict
}
