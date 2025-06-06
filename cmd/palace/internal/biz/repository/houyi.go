package repository

import (
	"context"

	"github.com/aide-family/moon/pkg/api/common"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
)

type Houyi interface {
	Sync() (HouyiSyncClient, bool)
	Query() (HouyiQueryClient, bool)
}

type HouyiSyncClient interface {
	SyncMetricMetadata(ctx context.Context, req *houyiv1.MetricMetadataRequest) (*houyiv1.SyncReply, error)
	SyncMetricDatasource(ctx context.Context, req *houyiv1.MetricDatasourceRequest) (*houyiv1.SyncReply, error)
	SyncMetricStrategy(ctx context.Context, req *houyiv1.MetricStrategyRequest) (*houyiv1.SyncReply, error)
}

type HouyiQueryClient interface {
	MetricDatasourceQuery(ctx context.Context, req *houyiv1.MetricDatasourceQueryRequest) (*common.MetricDatasourceQueryReply, error)
}
