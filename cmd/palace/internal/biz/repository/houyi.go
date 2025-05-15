package repository

import (
	"context"

	"github.com/moon-monitor/moon/pkg/api/common"
	houyiv1 "github.com/moon-monitor/moon/pkg/api/houyi/v1"
)

type Houyi interface {
	Sync() (HouyiSyncClient, bool)
	Query() (HouyiQueryClient, bool)
}

type HouyiSyncClient interface {
	MetricMetadata(ctx context.Context, req *houyiv1.MetricMetadataRequest) (*houyiv1.SyncReply, error)
}

type HouyiQueryClient interface {
	MetricDatasourceQuery(ctx context.Context, req *houyiv1.MetricDatasourceQueryRequest) (*common.MetricDatasourceQueryReply, error)
}
