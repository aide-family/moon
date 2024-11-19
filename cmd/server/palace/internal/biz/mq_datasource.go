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

func NewMqDataSourceBiz(datasourceRepo repository.MQDataSource) *MqDataSourceBiz {
	return &MqDataSourceBiz{
		datasourceRepo: datasourceRepo,
	}
}

// GetMqDataSource get mq datasource detail
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

// CreateMqDataSource create mq datasource
func (m *MqDataSourceBiz) CreateMqDataSource(ctx context.Context, params *bo.CreateMqDatasourceParams) error {
	if err := m.datasourceRepo.CreateMqDatasource(ctx, params); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateMqDataSource update mq datasource
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

// UpdateMqDataSourceStatus update mq datasource status
func (m *MqDataSourceBiz) UpdateMqDataSourceStatus(ctx context.Context, params *bo.UpdateMqDatasourceStatusParams) error {
	if err := m.datasourceRepo.UpdateMqDatasourceStatus(ctx, params); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// MqDataSourceList mq datasource list
func (m *MqDataSourceBiz) MqDataSourceList(ctx context.Context, params *bo.QueryMqDatasourceListParams) ([]*bizmodel.MqDatasource, error) {
	datasource, err := m.datasourceRepo.ListMqDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return datasource, nil
}

// GetMqDatasourceSelect mq datasource select
func (m *MqDataSourceBiz) GetMqDatasourceSelect(ctx context.Context, param *bo.QueryMqDatasourceListParams) ([]*bizmodel.MqDatasource, error) {
	list, err := m.datasourceRepo.ListMqDatasource(ctx, param)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return list, nil
}

func (m *MqDataSourceBiz) DeleteMqDatasource(ctx context.Context, ID uint32) error {
	if err := m.datasourceRepo.DeleteMqDatasource(ctx, ID); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}
