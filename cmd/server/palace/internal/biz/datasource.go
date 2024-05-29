package biz

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"github.com/aide-cloud/moon/pkg/types"
	"github.com/aide-cloud/moon/pkg/vobj"
)

func NewDatasourceBiz(datasourceRepository repository.Datasource) *DatasourceBiz {
	return &DatasourceBiz{
		datasourceRepository: datasourceRepository,
	}
}

// DatasourceBiz .
type DatasourceBiz struct {
	datasourceRepository repository.Datasource
}

// CreateDatasource 创建数据源
func (b *DatasourceBiz) CreateDatasource(ctx context.Context, datasource *bo.CreateDatasourceParams) (*bizmodel.Datasource, error) {
	return b.datasourceRepository.CreateDatasource(ctx, datasource)
}

// UpdateDatasourceBaseInfo 更新数据源
func (b *DatasourceBiz) UpdateDatasourceBaseInfo(ctx context.Context, datasource *bo.UpdateDatasourceBaseInfoParams) error {
	return b.datasourceRepository.UpdateDatasourceBaseInfo(ctx, datasource)
}

// UpdateDatasourceConfig 更新数据源配置
func (b *DatasourceBiz) UpdateDatasourceConfig(ctx context.Context, datasource *bo.UpdateDatasourceConfigParams) error {
	return b.datasourceRepository.UpdateDatasourceConfig(ctx, datasource)
}

// ListDatasource 获取数据源列表
func (b *DatasourceBiz) ListDatasource(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error) {
	return b.datasourceRepository.ListDatasource(ctx, params)
}

// DeleteDatasource 删除数据源
func (b *DatasourceBiz) DeleteDatasource(ctx context.Context, id uint32) error {
	return b.datasourceRepository.DeleteDatasourceByID(ctx, id)
}

// GetDatasource 获取数据源详情
func (b *DatasourceBiz) GetDatasource(ctx context.Context, id uint32) (*bizmodel.Datasource, error) {
	return b.datasourceRepository.GetDatasource(ctx, id)
}

// GetDatasourceSelect 获取数据源下拉列表
func (b *DatasourceBiz) GetDatasourceSelect(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bo.SelectOptionBo, error) {
	list, err := b.datasourceRepository.ListDatasource(ctx, params)
	if err != nil {
		return nil, err
	}
	return types.SliceTo(list, func(item *bizmodel.Datasource) *bo.SelectOptionBo {
		return bo.NewDatasourceOptionBuild(item).ToSelectOption()
	}), nil
}

// UpdateDatasourceStatus 更新数据源状态
func (b *DatasourceBiz) UpdateDatasourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	// 需要校验数据源是否被使用， 是否有权限
	return b.datasourceRepository.UpdateDatasourceStatus(ctx, status, ids...)
}
