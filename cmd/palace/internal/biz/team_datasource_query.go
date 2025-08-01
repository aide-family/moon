package biz

import (
	"context"

	"github.com/aide-family/moon/cmd/palace/internal/biz/bo"
	"github.com/aide-family/moon/cmd/palace/internal/biz/repository"
	"github.com/aide-family/moon/pkg/api/common"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
	"github.com/aide-family/moon/pkg/plugin/datasource"
	"github.com/go-kratos/kratos/v2/log"
)

func NewTeamDatasourceQueryBiz(
	teamDatasourceMetricRepo repository.TeamDatasourceMetric,
	houyiRepo repository.Houyi,
	logger log.Logger,
) *TeamDatasourceQuery {
	return &TeamDatasourceQuery{
		teamDatasourceMetricRepo: teamDatasourceMetricRepo,
		houyiRepo:                houyiRepo,
		helper:                   log.NewHelper(logger),
	}
}

type TeamDatasourceQuery struct {
	teamDatasourceMetricRepo repository.TeamDatasourceMetric
	houyiRepo                repository.Houyi
	helper                   *log.Helper
}

func (t *TeamDatasourceQuery) MetricDatasourceQuery(ctx context.Context, req *bo.MetricDatasourceQueryRequest) (*common.MetricDatasourceQueryReply, error) {
	queryClient, ok := t.houyiRepo.Query()
	if !ok {
		return t.metricDatasourceQuery(ctx, req)
	}
	params := &houyiv1.MetricDatasourceQueryRequest{
		Datasource: bo.ToSyncMetricDatasourceItem(req.Datasource),
		Expr:       req.Expr,
		Time:       req.Time,
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		Step:       req.Step,
	}
	return queryClient.MetricDatasourceQuery(ctx, params)
}

func (t *TeamDatasourceQuery) metricDatasourceQuery(ctx context.Context, req *bo.MetricDatasourceQueryRequest) (*common.MetricDatasourceQueryReply, error) {
	datasourceInstance, err := bo.ToMetricDatasource(req.Datasource, t.helper.Logger())
	if err != nil {
		return nil, err
	}
	queryParams := &datasource.MetricQueryRequest{
		Expr:      req.Expr,
		Time:      req.Time,
		EndTime:   req.EndTime,
		Step:      req.Step,
		StartTime: req.StartTime,
	}

	return bo.ToMetricDatasourceQueryReply(datasourceInstance.Query(ctx, queryParams))
}
