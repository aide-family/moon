package repoimpl

import (
	"context"

	"gorm.io/gen"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/helper/middleware"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel/bizquery"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
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
		return nil, bo.UnLoginErr
	}
	bizDB, err := data.GetBizGormDB(claims.GetTeam())
	if err != nil {
		return nil, err
	}
	return bizquery.Use(bizDB), nil
}

func (l *datasourceRepositoryImpl) CreateDatasource(ctx context.Context, datasource *bo.CreateDatasourceParams) (*bizmodel.Datasource, error) {
	q, err := getBizDB(ctx, l.data)
	if err != nil {
		return nil, err
	}
	datasourceModel := &bizmodel.Datasource{
		Name:     datasource.Name,
		Category: datasource.Type,
		Config:   datasource.Config,
		Endpoint: datasource.Endpoint,
		Status:   datasource.Status,
		Remark:   datasource.Remark,
	}
	if err = q.Datasource.WithContext(ctx).Create(datasourceModel); err != nil {
		return nil, err
	}
	return datasourceModel, nil
}

func (l *datasourceRepositoryImpl) GetDatasource(ctx context.Context, id uint32) (*bizmodel.Datasource, error) {
	q, err := getBizDB(ctx, l.data)
	if err != nil {
		return nil, err
	}
	return q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(id)).First()
}

func (l *datasourceRepositoryImpl) ListDatasource(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error) {
	q, err := getBizDB(ctx, l.data)
	if err != nil {
		return nil, err
	}
	qq := q.Datasource.WithContext(ctx)
	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, q.Datasource.Name.Like(params.Keyword))
	}
	if !params.Type.IsUnknown() {
		wheres = append(wheres, q.Datasource.Category.Eq(params.Type.GetValue()))
	}
	if !params.Status.IsUnknown() {
		wheres = append(wheres, q.Datasource.Status.Eq(params.Status.GetValue()))
	}
	if !types.IsNil(params.Page) {
		page := params.Page
		total, err := qq.Count()
		if err != nil {
			return nil, err
		}
		params.Page.SetTotal(int(total))
		pageNum, pageSize := page.GetPageNum(), page.GetPageSize()
		if pageNum <= 1 {
			qq = qq.Limit(pageSize)
		} else {
			qq = qq.Offset((pageNum - 1) * pageSize).Limit(pageSize)
		}
	}
	return qq.Where(wheres...).Find()
}

func (l *datasourceRepositoryImpl) UpdateDatasourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	q, err := getBizDB(ctx, l.data)
	if err != nil {
		return err
	}
	_, err = q.Datasource.WithContext(ctx).Where(q.Datasource.ID.In(ids...)).Update(q.Datasource.Status, status)
	return err
}

func (l *datasourceRepositoryImpl) UpdateDatasourceBaseInfo(ctx context.Context, datasource *bo.UpdateDatasourceBaseInfoParams) error {
	q, err := getBizDB(ctx, l.data)
	if err != nil {
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
	if err != nil {
		return err
	}
	_, err = q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(datasource.ID)).UpdateColumnSimple(
		q.Datasource.Config.Value(datasource.Config),
		q.Datasource.Category.Value(datasource.Type.GetValue()),
	)
	return err
}

func (l *datasourceRepositoryImpl) DeleteDatasourceByID(ctx context.Context, id uint32) error {
	q, err := getBizDB(ctx, l.data)
	if err != nil {
		return err
	}
	_, err = q.Datasource.WithContext(ctx).Where(q.Datasource.ID.Eq(id)).Delete()
	return err
}
