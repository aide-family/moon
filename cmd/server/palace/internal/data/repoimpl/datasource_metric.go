package repoimpl

import (
	"context"

	"github.com/aide-family/moon/cmd/server/palace/internal/biz/repository"
	"github.com/aide-family/moon/cmd/server/palace/internal/data"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel"
	"github.com/aide-family/moon/pkg/palace/model/bizmodel/bizquery"
	"github.com/aide-family/moon/pkg/util/types"

	"gorm.io/gorm/clause"
)

// NewDatasourceMetricRepository 创建数据源指标仓库
func NewDatasourceMetricRepository(data *data.Data) repository.DatasourceMetric {
	return &datasourceMetricRepositoryImpl{data: data}
}

type datasourceMetricRepositoryImpl struct {
	data *data.Data
}

func (l *datasourceMetricRepositoryImpl) CreateMetrics(ctx context.Context, metrics ...*bizmodel.DatasourceMetric) error {
	q, err := getBizDB(ctx, l.data)
	if !types.IsNil(err) {
		return err
	}
	return q.DatasourceMetric.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(metrics, 10)
}

func (l *datasourceMetricRepositoryImpl) CreateMetricsNoAuth(ctx context.Context, teamID uint32, metrics ...*bizmodel.DatasourceMetric) error {
	bizDB, err := l.data.GetBizGormDB(teamID)
	if !types.IsNil(err) {
		return err
	}
	q := bizquery.Use(bizDB)
	return q.DatasourceMetric.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(metrics, 10)
}
