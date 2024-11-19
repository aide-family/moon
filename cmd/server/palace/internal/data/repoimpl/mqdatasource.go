package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gen"
)

// NewMqDatasourceRepository 创建mq数据源
func NewMqDatasourceRepository(data *data.Data) repository.MQDataSource {
	return &mqDatasourceRepositoryImpl{data: data}
}

type mqDatasourceRepositoryImpl struct {
	data *data.Data
}

func (m *mqDatasourceRepositoryImpl) CreateMqDatasource(ctx context.Context, params *bo.CreateMqDatasourceParams) error {
	bizQuery, err := getBizQuery(ctx, m.data)
	if !types.IsNil(err) {
		return err
	}
	datasourceModel := createMqDatasourceToModel(ctx, params)
	if err := bizQuery.MqDatasource.WithContext(ctx).Create(datasourceModel); !types.IsNil(err) {
		return err
	}
	return nil
}

func (m *mqDatasourceRepositoryImpl) DeleteMqDatasource(ctx context.Context, ID uint32) error {
	bizQuery, err := getBizQuery(ctx, m.data)
	if !types.IsNil(err) {
		return err
	}
	if _, err := bizQuery.MqDatasource.WithContext(ctx).Where(bizQuery.MqDatasource.ID.Eq(ID)).Delete(); !types.IsNil(err) {
		return err
	}
	return nil
}

func (m *mqDatasourceRepositoryImpl) UpdateMqDatasource(ctx context.Context, params *bo.UpdateMqDatasourceParams) error {
	bizQuery, err := getBizQuery(ctx, m.data)
	if !types.IsNil(err) {
		return err
	}
	updateModel := createMqDatasourceToModel(ctx, params.UpdateParam)
	if _, err := bizQuery.MqDatasource.WithContext(ctx).Where(bizQuery.MqDatasource.ID.Eq(params.ID)).
		Updates(updateModel); !types.IsNil(err) {
		return err
	}
	return nil
}

func (m *mqDatasourceRepositoryImpl) GetMqDatasource(ctx context.Context, id uint32) (*bizmodel.MqDatasource, error) {
	bizQuery, err := getBizQuery(ctx, m.data)
	if !types.IsNil(err) {
		return nil, err
	}
	return bizQuery.MqDatasource.WithContext(ctx).Where(bizQuery.MqDatasource.ID.Eq(id)).First()
}

func (m *mqDatasourceRepositoryImpl) ListMqDatasource(ctx context.Context, params *bo.QueryMqDatasourceListParams) ([]*bizmodel.MqDatasource, error) {
	bizQuery, err := getBizQuery(ctx, m.data)
	if !types.IsNil(err) {
		return nil, err
	}

	var wheres []gen.Condition
	if !types.TextIsNull(params.Keyword) {
		wheres = append(wheres, bizQuery.MqDatasource.Name.Like(params.Keyword))
	}

	if !params.DatasourceType.IsUnknown() {
		wheres = append(wheres, bizQuery.MqDatasource.DatasourceType.Eq(params.DatasourceType.GetValue()))
	}

	if !params.StorageType.IsUnknown() {
		wheres = append(wheres, bizQuery.MqDatasource.StorageType.Eq(params.StorageType.GetValue()))
	}

	if !params.Status.IsUnknown() {
		wheres = append(wheres, bizQuery.MqDatasource.Status.Eq(params.Status.GetValue()))
	}
	bizWrapper := bizQuery.MqDatasource.WithContext(ctx).Where(wheres...)

	if bizWrapper, err = types.WithPageQuery(bizWrapper, params.Page); err != nil {
		return nil, err
	}

	return bizWrapper.Order(bizQuery.MqDatasource.ID.Desc()).Find()
}

func (m *mqDatasourceRepositoryImpl) UpdateMqDatasourceStatus(ctx context.Context, params *bo.UpdateMqDatasourceStatusParams) error {
	bizQuery, err := getBizQuery(ctx, m.data)
	if !types.IsNil(err) {
		return err
	}
	if _, err := bizQuery.MqDatasource.WithContext(ctx).Where(bizQuery.MqDatasource.ID.Eq(params.ID)).
		UpdateColumn(bizQuery.MqDatasource.Status, params.Status); !types.IsNil(err) {
		return err
	}
	return nil
}

func createMqDatasourceToModel(ctx context.Context, params *bo.CreateMqDatasourceParams) *bizmodel.MqDatasource {
	config, _ := types.Marshal(params.Config)
	datasourceModel := &bizmodel.MqDatasource{
		Name:           params.Name,
		DatasourceType: params.DatasourceType,
		StorageType:    params.StorageType,
		Config:         string(config),
		Endpoint:       params.Endpoint,
		Status:         params.Status,
		Remark:         params.Remark,
	}
	datasourceModel.WithContext(ctx)
	return datasourceModel
}
