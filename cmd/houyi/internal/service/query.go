package service

import (
	"context"

	"github.com/aide-family/moon/cmd/houyi/internal/biz"
	"github.com/aide-family/moon/cmd/houyi/internal/biz/bo"
	"github.com/aide-family/moon/cmd/houyi/internal/service/build"
	"github.com/aide-family/moon/pkg/api/common"
	houyiv1 "github.com/aide-family/moon/pkg/api/houyi/v1"
)

type QueryService struct {
	houyiv1.UnimplementedQueryServer

	metric *biz.Metric
}

func NewQueryService(metric *biz.Metric) *QueryService {
	return &QueryService{
		metric: metric,
	}
}

func (s *QueryService) MetricDatasourceQuery(ctx context.Context, req *houyiv1.MetricDatasourceQueryRequest) (*common.MetricDatasourceQueryReply, error) {
	datasourceConfig, err := build.ToMetricDatasourceConfig(req.GetDatasource())
	if err != nil {
		return nil, err
	}
	params := &bo.MetricDatasourceQueryRequest{
		Datasource: datasourceConfig,
		Expr:       req.GetExpr(),
		Time:       req.GetTime(),
		StartTime:  req.GetStartTime(),
		EndTime:    req.GetEndTime(),
		Step:       req.GetStep(),
	}
	reply, err := s.metric.QueryMetricDatasource(ctx, params)
	if err != nil {
		return nil, err
	}
	return &common.MetricDatasourceQueryReply{
		Results: build.ToMetricQueryResults(reply.Results),
	}, nil
}
