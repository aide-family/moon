package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/datasource"
)

type MetricQueryService struct {
	v1.UnimplementedMetricQueryServiceServer
	gateway *datasource.Gateway
}

func NewMetricQueryService(gateway *datasource.Gateway) *MetricQueryService {
	return &MetricQueryService{gateway: gateway}
}

func (s *MetricQueryService) QueryMetric(ctx context.Context, req *v1.QueryMetricRequest) (*v1.QueryMetricResponse, error) {
	body, err := s.gateway.Query(ctx, req.GetDatasourceUid(), req.GetExpr(), req.GetTime())
	if err != nil {
		return nil, err
	}
	return &v1.QueryMetricResponse{Body: body}, nil
}

func (s *MetricQueryService) QueryMetricRange(ctx context.Context, req *v1.QueryMetricRangeRequest) (*v1.QueryMetricRangeResponse, error) {
	body, err := s.gateway.QueryRange(ctx, req.GetDatasourceUid(), req.GetExpr(), req.GetStart(), req.GetEnd(), req.GetStep())
	if err != nil {
		return nil, err
	}
	return &v1.QueryMetricRangeResponse{Body: body}, nil
}

func (s *MetricQueryService) ProxyQuery(ctx context.Context, req *v1.ProxyQueryRequest) (*v1.ProxyQueryResponse, error) {
	body, err := s.gateway.Proxy(ctx, req.GetDatasourceUid(), req.GetMethod(), req.GetPath(), req.GetBody())
	if err != nil {
		return nil, err
	}
	return &v1.ProxyQueryResponse{Body: body}, nil
}
