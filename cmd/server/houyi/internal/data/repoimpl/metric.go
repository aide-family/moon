package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/bo"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/biz/repository"
	"github.com/aide-cloud/moon/cmd/server/houyi/internal/data"
)

func NewMetricRepository(data *data.Data) repository.Metric {
	return &metricRepositoryImpl{data: data}
}

type metricRepositoryImpl struct {
	data *data.Data
}

func (l *metricRepositoryImpl) GetMetrics(ctx context.Context, datasourceInfo *bo.GetMetricsParams) ([]*bo.MetricDetail, error) {
	//TODO implement me
	panic("implement me")
}
