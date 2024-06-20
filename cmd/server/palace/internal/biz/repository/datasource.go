package repository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/vobj"
)

// Datasource .
type Datasource interface {
	// CreateDatasource 创建数据源
	CreateDatasource(ctx context.Context, datasource *bo.CreateDatasourceParams) (*bizmodel.Datasource, error)

	// GetDatasource 获取数据源详情
	GetDatasource(ctx context.Context, id uint32) (*bizmodel.Datasource, error)

	// GetDatasourceNoAuth 获取数据源详情(不鉴权)
	GetDatasourceNoAuth(ctx context.Context, id, teamId uint32) (*bizmodel.Datasource, error)

	// ListDatasource 获取数据源列表
	ListDatasource(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error)

	// UpdateDatasourceStatus 更新数据源状态
	UpdateDatasourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error

	// UpdateDatasourceBaseInfo 更新数据源基础信息
	UpdateDatasourceBaseInfo(ctx context.Context, datasource *bo.UpdateDatasourceBaseInfoParams) error

	// UpdateDatasourceConfig 更新数据源配置
	UpdateDatasourceConfig(ctx context.Context, datasource *bo.UpdateDatasourceConfigParams) error

	// DeleteDatasourceByID 删除数据源
	DeleteDatasourceByID(ctx context.Context, id uint32) error
}
