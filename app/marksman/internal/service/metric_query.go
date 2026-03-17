package service

import (
	"context"
	"encoding/json"

	"github.com/bwmarrin/snowflake"
	"google.golang.org/protobuf/types/known/structpb"

	"github.com/aide-family/marksman/internal/biz"
	apiv1 "github.com/aide-family/marksman/pkg/api/v1"
)

// NewMetricQueryService creates the MetricQuery HTTP/gRPC service.
func NewMetricQueryService(metricQueryBiz *biz.MetricQueryBiz) *MetricQueryService {
	return &MetricQueryService{
		metricQueryBiz: metricQueryBiz,
	}
}

// MetricQueryService implements apiv1.MetricQueryServer and MetricQueryHTTPServer.
type MetricQueryService struct {
	apiv1.UnimplementedMetricQueryServer

	metricQueryBiz *biz.MetricQueryBiz
}

// Query runs an instant query and returns the response as a structured any type.
func (s *MetricQueryService) Query(ctx context.Context, req *apiv1.MetricQueryRequest) (*apiv1.MetricQueryReply, error) {
	jsonStr, err := s.metricQueryBiz.Query(ctx, snowflake.ParseInt64(req.GetUid()), req.GetQuery(), req.GetTime())
	if err != nil {
		return nil, err
	}
	result, err := jsonToStruct(jsonStr)
	if err != nil {
		return nil, err
	}
	return &apiv1.MetricQueryReply{Response: result}, nil
}

// QueryRange runs a range query and returns the response as a structured any type.
func (s *MetricQueryService) QueryRange(ctx context.Context, req *apiv1.MetricQueryRangeRequest) (*apiv1.MetricQueryRangeReply, error) {
	jsonStr, err := s.metricQueryBiz.QueryRange(ctx, snowflake.ParseInt64(req.GetUid()), req.GetQuery(), req.GetStart(), req.GetEnd(), req.GetStep())
	if err != nil {
		return nil, err
	}
	result, err := jsonToStruct(jsonStr)
	if err != nil {
		return nil, err
	}
	return &apiv1.MetricQueryRangeReply{Response: result}, nil
}

// Proxy forwards the request to the datasource and returns status code and body.
func (s *MetricQueryService) Proxy(ctx context.Context, req *apiv1.MetricQueryProxyRequest) (*apiv1.MetricQueryProxyReply, error) {
	statusCode, body, err := s.metricQueryBiz.Proxy(ctx, snowflake.ParseInt64(req.GetUid()), req.GetPath(), req.GetMethod(), req.GetBody())
	if err != nil {
		return nil, err
	}
	result, err := jsonToStruct(string(body))
	if err != nil {
		return nil, err
	}
	return &apiv1.MetricQueryProxyReply{StatusCode: int32(statusCode), Response: result}, nil
}

// jsonToStruct parses a JSON string into a google.protobuf.Struct (any type).
func jsonToStruct(jsonStr string) (*structpb.Struct, error) {
	var m map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &m); err != nil {
		return nil, err
	}
	return structpb.NewStruct(m)
}
