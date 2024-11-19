package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

type (
	MQDataSource interface {
		// CreateMqDatasource 创建MQ数据源
		CreateMqDatasource(ctx context.Context, params *bo.CreateMqDatasourceParams) error
		// DeleteMqDatasource 删除MQ数据源
		DeleteMqDatasource(ctx context.Context, ID uint32) error
		// UpdateMqDatasource 更新MQ数据源
		UpdateMqDatasource(ctx context.Context, params *bo.UpdateMqDatasourceParams) error
		// GetMqDatasource 获取MQ数据源
		GetMqDatasource(ctx context.Context, id uint32) (*bizmodel.MqDatasource, error)
		// ListMqDatasource 获取MQ数据源列表
		ListMqDatasource(ctx context.Context, params *bo.QueryMqDatasourceListParams) ([]*bizmodel.MqDatasource, error)
		// UpdateMqDatasourceStatus 更新MQ数据源状态
		UpdateMqDatasourceStatus(ctx context.Context, params *bo.UpdateMqDatasourceStatusParams) error
	}
)
