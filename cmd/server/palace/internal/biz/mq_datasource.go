package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// MqDataSourceBiz mq datasource biz
type MqDataSourceBiz struct {
	datasourceRepo repository.MQDataSource
}

// NewMqDataSourceBiz 创建 mq 数据源业务
func NewMqDataSourceBiz(datasourceRepo repository.MQDataSource) *MqDataSourceBiz {
	return &MqDataSourceBiz{
		datasourceRepo: datasourceRepo,
	}
}

// GetMqDataSource 获取 mq 数据源详情
func (m *MqDataSourceBiz) GetMqDataSource(ctx context.Context, ID uint32) (*bizmodel.MqDatasource, error) {
	detail, err := m.datasourceRepo.GetMqDatasource(ctx, ID)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastDataSourceNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return detail, nil
}

// CreateMqDataSource 创建 mq 数据源
func (m *MqDataSourceBiz) CreateMqDataSource(ctx context.Context, params *bo.CreateMqDatasourceParams) error {
	if err := m.datasourceRepo.CreateMqDatasource(ctx, params); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateMqDataSource 更新 mq 数据源
func (m *MqDataSourceBiz) UpdateMqDataSource(ctx context.Context, param *bo.UpdateMqDatasourceParams) error {
	// 参数校验
	if param.UpdateParam == nil {
		return merr.ErrorI18nParameterRelatedUpdateParameterNotFound(ctx)
	}
	if err := m.datasourceRepo.UpdateMqDatasource(ctx, param); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateMqDataSourceStatus 更新 mq 数据源状态
func (m *MqDataSourceBiz) UpdateMqDataSourceStatus(ctx context.Context, params *bo.UpdateMqDatasourceStatusParams) error {
	if err := m.datasourceRepo.UpdateMqDatasourceStatus(ctx, params); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// MqDataSourceList 获取 mq 数据源列表
func (m *MqDataSourceBiz) MqDataSourceList(ctx context.Context, params *bo.QueryMqDatasourceListParams) ([]*bizmodel.MqDatasource, error) {
	datasource, err := m.datasourceRepo.ListMqDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return datasource, nil
}

// GetMqDatasourceSelect 获取 mq 数据源选择列表
func (m *MqDataSourceBiz) GetMqDatasourceSelect(ctx context.Context, param *bo.QueryMqDatasourceListParams) ([]*bizmodel.MqDatasource, error) {
	list, err := m.datasourceRepo.ListMqDatasource(ctx, param)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return list, nil
}

// DeleteMqDatasource 删除 mq 数据源
func (m *MqDataSourceBiz) DeleteMqDatasource(ctx context.Context, id uint32) error {
	if err := m.datasourceRepo.DeleteMqDatasource(ctx, id); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}
