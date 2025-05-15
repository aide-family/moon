package repository

import (
	"context"

	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/bo"
	"github.com/moon-monitor/moon/cmd/houyi/internal/biz/do"
)

type MetricInit interface {
	Init(config bo.MetricDatasourceConfig) (Metric, error)
}

type Metric interface {
	Query(ctx context.Context, req *bo.MetricQueryRequest) ([]*do.MetricQueryReply, error)

	QueryRange(ctx context.Context, req *bo.MetricRangeQueryRequest) ([]*do.MetricQueryRangeReply, error)

	Metadata(ctx context.Context) (<-chan []*do.MetricItem, error)
}
