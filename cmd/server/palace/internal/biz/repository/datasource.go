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
	CreateDatasource(context.Context, *bo.CreateDatasourceParams) (*bizmodel.Datasource, error)

	// GetDatasource 获取数据源详情
	GetDatasource(context.Context, uint32) (*bizmodel.Datasource, error)

	// GetDatasourceNoAuth 获取数据源详情(不鉴权)
	GetDatasourceNoAuth(context.Context, uint32, uint32) (*bizmodel.Datasource, error)

	// ListDatasource 获取数据源列表
	ListDatasource(context.Context, *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error)

	// UpdateDatasourceStatus 更新数据源状态
	UpdateDatasourceStatus(context.Context, vobj.Status, ...uint32) error

	// UpdateDatasourceBaseInfo 更新数据源基础信息
	UpdateDatasourceBaseInfo(context.Context, *bo.UpdateDatasourceBaseInfoParams) error

	// UpdateDatasourceConfig 更新数据源配置
	UpdateDatasourceConfig(context.Context, *bo.UpdateDatasourceConfigParams) error

	// DeleteDatasourceByID 删除数据源
	DeleteDatasourceByID(context.Context, uint32) error
}
