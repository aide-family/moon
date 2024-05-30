package repoimpl

import (
	"context"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/data"
	"github.com/aide-cloud/moon/pkg/helper/model/bizmodel"
	"gorm.io/gorm/clause"
)

func NewDatasourceMetricRepository(data *data.Data) repository.DatasourceMetric {
	return &datasourceMetricRepositoryImpl{data: data}
}

type datasourceMetricRepositoryImpl struct {
	data *data.Data
}

func (l *datasourceMetricRepositoryImpl) CreateMetrics(ctx context.Context, metrics ...*bizmodel.DatasourceMetric) error {
	q, err := getBizDB(ctx, l.data)
	if err != nil {
		return err
	}
	return q.DatasourceMetric.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(metrics...)
}
