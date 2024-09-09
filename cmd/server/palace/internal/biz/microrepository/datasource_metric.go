package microrepository

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/bo"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
)

// DatasourceMetric .
type DatasourceMetric interface {
	// GetMetadata 同步指标元数据
	GetMetadata(ctx context.Context, datasourceInfo *bizmodel.Datasource) ([]*bizmodel.DatasourceMetric, error)

	// InitiateSyncRequest 发起同步请求
	InitiateSyncRequest(ctx context.Context, datasourceInfo *bizmodel.Datasource) error

	// Query 查询指标数据
	Query(ctx context.Context, req *bo.DatasourceQueryParams) ([]*bo.MetricQueryData, error)
}
