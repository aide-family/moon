package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/microrepository"
	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/helper/middleware"
	"github.com/aide-family/moon/pkg/merr"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"

	"github.com/go-kratos/kratos/v2/errors"
	"gorm.io/gorm"
)

// NewDatasourceBiz 创建数据源业务
func NewDatasourceBiz(
	datasourceRepository repository.Datasource,
	datasourceMetricRepository repository.DatasourceMetric,
	datasourceMetricMicroRepository microrepository.DatasourceMetric,
	lock repository.Lock,
) *DatasourceBiz {
	return &DatasourceBiz{
		lock:                            lock,
		datasourceRepository:            datasourceRepository,
		datasourceMetricRepository:      datasourceMetricRepository,
		datasourceMetricMicroRepository: datasourceMetricMicroRepository,
	}
}

// DatasourceBiz 数据源业务
type DatasourceBiz struct {
	lock                            repository.Lock
	datasourceRepository            repository.Datasource
	datasourceMetricRepository      repository.DatasourceMetric
	datasourceMetricMicroRepository microrepository.DatasourceMetric
}

// CreateDatasource 创建数据源
func (b *DatasourceBiz) CreateDatasource(ctx context.Context, datasource *bo.CreateDatasourceParams) (*bizmodel.Datasource, error) {
	datasourceDo, err := b.datasourceRepository.CreateDatasource(ctx, datasource)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return datasourceDo, err
}

// UpdateDatasourceBaseInfo 更新数据源
func (b *DatasourceBiz) UpdateDatasourceBaseInfo(ctx context.Context, datasource *bo.UpdateDatasourceBaseInfoParams) error {
	// 校验config数据格式
	if datasource.ConfigValue != "" {
		_, err := types.JSONToMap(datasource.ConfigValue)
		if err != nil {
			return merr.ErrorI18nNotificationDataSourceConfigurationError(ctx).WithCause(err)
		}
	}
	if err := b.datasourceRepository.UpdateDatasourceBaseInfo(ctx, datasource); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// UpdateDatasourceConfig 更新数据源配置
func (b *DatasourceBiz) UpdateDatasourceConfig(ctx context.Context, datasource *bo.UpdateDatasourceConfigParams) error {
	if err := b.datasourceRepository.UpdateDatasourceConfig(ctx, datasource); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// ListDatasource 获取数据源列表
func (b *DatasourceBiz) ListDatasource(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bizmodel.Datasource, error) {
	list, err := b.datasourceRepository.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return list, nil
}

// DeleteDatasource 删除数据源
func (b *DatasourceBiz) DeleteDatasource(ctx context.Context, id uint32) error {
	if err := b.datasourceRepository.DeleteDatasourceByID(ctx, id); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// GetDatasource 获取数据源详情
func (b *DatasourceBiz) GetDatasource(ctx context.Context, id uint32) (*bizmodel.Datasource, error) {
	detail, err := b.datasourceRepository.GetDatasource(ctx, id)
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, merr.ErrorI18nToastDataSourceNotFound(ctx)
		}
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return detail, nil
}

// GetDatasourceSelect 获取数据源下拉列表
func (b *DatasourceBiz) GetDatasourceSelect(ctx context.Context, params *bo.QueryDatasourceListParams) ([]*bo.SelectOptionBo, error) {
	list, err := b.datasourceRepository.ListDatasource(ctx, params)
	if !types.IsNil(err) {
		return nil, merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return types.SliceTo(list, func(item *bizmodel.Datasource) *bo.SelectOptionBo {
		return bo.NewDatasourceOptionBuild(item).ToSelectOption()
	}), nil
}

// UpdateDatasourceStatus 更新数据源状态
func (b *DatasourceBiz) UpdateDatasourceStatus(ctx context.Context, status vobj.Status, ids ...uint32) error {
	// TODO 需要校验数据源是否被使用， 是否有权限
	if err := b.datasourceRepository.UpdateDatasourceStatus(ctx, status, ids...); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

func syncDatasourceMetaKey(datasourceID uint32) string {
	return fmt.Sprintf("sync:datasource:meta:%d", datasourceID)
}

// SyncDatasourceMetaV2 同步数据源元数据
func (b *DatasourceBiz) SyncDatasourceMetaV2(ctx context.Context, id uint32) (err error) {
	if err := b.lock.Lock(ctx, syncDatasourceMetaKey(id), 10*time.Minute); !types.IsNil(err) {
		return merr.ErrorI18nToastDatasourceSyncing(ctx).WithCause(err)
	}
	defer func() {
		if err != nil {
			_ = b.lock.UnLock(context.Background(), syncDatasourceMetaKey(id))
		}
	}()

	// 获取数据源详情
	datasourceDetail, err := b.datasourceRepository.GetDatasourceNoAuth(ctx, id, middleware.GetTeamID(ctx))
	if !types.IsNil(err) {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return merr.ErrorI18nToastDataSourceNotFound(ctx)
		}
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	if err := b.datasourceMetricMicroRepository.InitiateSyncRequest(ctx, datasourceDetail); !types.IsNil(err) {
		return merr.ErrorI18nNotificationSystemError(ctx).WithCause(err)
	}
	return nil
}

// Query 查询数据
func (b *DatasourceBiz) Query(ctx context.Context, params *bo.DatasourceQueryParams) ([]*bo.MetricQueryData, error) {
	if types.IsNil(params.Datasource) {
		// 查询数据源
		datasourceDetail, err := b.datasourceRepository.GetDatasource(ctx, params.DatasourceID)
		if !types.IsNil(err) {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, merr.ErrorI18nToastDataSourceNotFound(ctx)
			}
			return nil, err
		}
		params.Datasource = datasourceDetail
	}

	return b.datasourceMetricMicroRepository.Query(ctx, params)
}
