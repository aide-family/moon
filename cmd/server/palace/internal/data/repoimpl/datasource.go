package repoimpl

import (
	"context"

	"github.com/aide-family/moon/api/merr"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"gorm.io/gen/field"

	"gorm.io/gen"
)

// NewDatasourceRepository 创建数据源
func NewDatasourceRepository(data *data.Data) repository.Datasource {
	return &datasourceRepositoryImpl{data: data}
}

type datasourceRepositoryImpl struct {
	data *data.Data
}

// getBizQuery 获取业务数据库
func getBizQuery(ctx context.Context, data *data.Data) (*bizquery.Query, error) {
	claims, ok := middleware.ParseJwtClaims(ctx)
	if !ok {
		return nil, merr.ErrorI18nUnLoginErr(ctx)
	}
	bizDB, err := data.GetBizGormDB(claims.GetTeam())
	if !types.IsNil(err) {
		return nil, err
	}
	return bizquery.Use(bizDB), nil
}

func (l *datasourceRepositoryImpl) CreateDatasource(ctx context.Context, datasource *bo.CreateDatasourceParams) (*bizmodel.Datasource, error) {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	datasourceModel := &bizmodel.Datasource{
		Name:        datasource.Name,
		Category:    datasource.Type,
		Config:      datasource.Config,
		Endpoint:    datasource.Endpoint,
		Status:      datasource.Status,
		Remark:      datasource.Remark,
		StorageType: datasource.StorageType,
	}
	datasourceModel.WithContext(ctx)
	if err = bizQuery.Datasource.WithContext(ctx).Create(datasourceModel); !types.IsNil(err) {
		return nil, err
	}
	return datasourceModel, nil
}

func (l *datasourceRepositoryImpl) GetDatasource(ctx context.Context, id uint32) (*bizmodel.Datasource, error) {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return bizQuery.Datasource.WithContext(ctx).Where(bizQuery.Datasource.ID.Eq(id)).Preload(field.Associations).First()
}

func (l *datasourceRepositoryImpl) GetDatasourceNoAuth(ctx context.Context, id, teamID uint32) (*bizmodel.Datasource, error) {
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return nil, err
	}
	bizQuery := bizquery.Use(bizDB)
	return bizQuery.Datasource.WithContext(ctx).Where(bizQuery.Datasource.ID.Eq(id)).First()
}

func (l *datasourceRepositoryImpl) ListDatasource(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error) {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	datasourcePreloadOp := bizQuery.Datasource.WithContext(ctx).Preload(field.Associations)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, bizQuery.Datasource.Name.Like(params.Keyword))
	}
	if !params.Type.IsUnknown() {
		wheres = append(wheres, bizQuery.Datasource.Category.Eq(params.Type.GetValue()))
	}
	if !params.StorageType.IsUnknown() {
		wheres = append(wheres, bizQuery.Datasource.StorageType.Eq(params.StorageType.GetValue()))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.Datasource.Status.Eq(params.Status.GetValue()))
	}
	datasourcePreloadOp = datasourcePreloadOp.Where(wheres...)
	if err := types.WithPageQuery[bizquery.IDatasourceDo](datasourcePreloadOp, params.Page); err != nil {
		return nil, err
	}
	return datasourcePreloadOp.Order(bizQuery.Datasource.ID.Desc()).Find()
}

func (l *datasourceRepositoryImpl) UpdateDatasourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = bizQuery.Datasource.WithContext(ctx).Where(bizQuery.Datasource.ID.In(ids...)).Update(bizQuery.Datasource.Status, status)
	return err
}

func (l *datasourceRepositoryImpl) UpdateDatasourceBaseInfo(ctx context.Context, datasource *bo.UpdateDatasourceBaseInfoParams) error {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = bizQuery.Datasource.WithContext(ctx).Where(bizQuery.Datasource.ID.Eq(datasource.ID)).UpdateColumnSimple(
		bizQuery.Datasource.Name.Value(datasource.Name),
		bizQuery.Datasource.Status.Value(datasource.Status.GetValue()),
		bizQuery.Datasource.Remark.Value(datasource.Remark),
	)
	return err
}

func (l *datasourceRepositoryImpl) UpdateDatasourceConfig(ctx context.Context, datasource *bo.UpdateDatasourceConfigParams) error {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = bizQuery.Datasource.WithContext(ctx).Where(bizQuery.Datasource.ID.Eq(datasource.ID)).UpdateColumnSimple(
		bizQuery.Datasource.Config.Value(datasource.Config),
		bizQuery.Datasource.Category.Value(datasource.Type.GetValue()),
		bizQuery.Datasource.StorageType.Value(datasource.StorageType.GetValue()),
	)
	return err
}

func (l *datasourceRepositoryImpl) DeleteDatasourceByID(ctx context.Context, id uint32) error {
	bizQuery, err := getBizQuery(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = bizQuery.Datasource.WithContext(ctx).Where(bizQuery.Datasource.ID.Eq(id)).Delete()
	return err
}
