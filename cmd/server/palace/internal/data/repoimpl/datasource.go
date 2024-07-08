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

func NewDatasourceRepository(data *data.Data) repository.Datasource {
	return &datasourceRepositoryImpl{data: data}
}

type datasourceRepositoryImpl struct {
	data *data.Data
}

// getBizDB 获取业务数据库
func getBizDB(ctx context.Context, data *data.Data) (*bizquery.Query, error) {
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
	q, err := getBizDB(ctx, l.data)
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
	if err = q.Datasource.WithContext(ctx).Create(datasourceModel); !types.IsNil(err) {
		return nil, err
	}
	return datasourceModel, nil
}

func (l *datasourceRepositoryImpl) GetDatasource(ctx context.Context, id uint32) (*bizmodel.Datasource, error) {
	q, err := getBizDB(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(id)).Preload(field.Associations).First()
}

func (l *datasourceRepositoryImpl) GetDatasourceNoAuth(ctx context.Context, id, teamId uint32) (*bizmodel.Datasource, error) {
	bizDB, err := l.data.GetBizGormDB(teamId)
	if !types.IsNil(err) {
		return nil, err
	}
	q := bizquery.Use(bizDB)
	return q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(id)).First()
}

func (l *datasourceRepositoryImpl) ListDatasource(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error) {
	q, err := getBizDB(ctx, l.data)
	if !types.IsNil(err) {
		return nil, err
	}
	qq := q.Datasource.WithContext(ctx).Preload(field.Associations)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, q.Datasource.Name.Like(params.Keyword))
	}
	if !params.Type.IsUnknown() {
		wheres = append(wheres, q.Datasource.Category.Eq(params.Type.GetValue()))
	}
	if !params.StorageType.IsUnknown() {
		wheres = append(wheres, q.Datasource.StorageType.Eq(params.StorageType.GetValue()))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, q.Datasource.Status.Eq(params.Status.GetValue()))
	}
	qq = qq.Where(wheres...)
	if err := types.WithPageQuery[bizquery.IDatasourceDo](qq, params.Page); err != nil {
		return nil, err
	}
	return qq.Order(bizquery.Datasource.ID.Desc()).Find()
}

func (l *datasourceRepositoryImpl) UpdateDatasourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	q, err := getBizDB(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = q.Datasource.WithContext(ctx).Where(q.Datasource.ID.In(ids...)).Update(q.Datasource.Status, status)
	return err
}

func (l *datasourceRepositoryImpl) UpdateDatasourceBaseInfo(ctx context.Context, datasource *bo.UpdateDatasourceBaseInfoParams) error {
	q, err := getBizDB(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(datasource.ID)).UpdateColumnSimple(
		q.Datasource.Name.Value(datasource.Name),
		q.Datasource.Status.Value(datasource.Status.GetValue()),
		q.Datasource.Remark.Value(datasource.Remark),
	)
	return err
}

func (l *datasourceRepositoryImpl) UpdateDatasourceConfig(ctx context.Context, datasource *bo.UpdateDatasourceConfigParams) error {
	q, err := getBizDB(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(datasource.ID)).UpdateColumnSimple(
		q.Datasource.Config.Value(datasource.Config),
		q.Datasource.Category.Value(datasource.Type.GetValue()),
		q.Datasource.StorageType.Value(datasource.StorageType.GetValue()),
	)
	return err
}

func (l *datasourceRepositoryImpl) DeleteDatasourceByID(ctx context.Context, id uint32) error {
	q, err := getBizDB(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	_, err = q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(id)).Delete()
	return err
}
