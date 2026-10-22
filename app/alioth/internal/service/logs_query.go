package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/datasource"
)

type LogsQueryService struct {
	v1.UnimplementedLogsQueryServiceServer
	gateway *datasource.Gateway
}

func NewLogsQueryService(gateway *datasource.Gateway) *LogsQueryService {
	return &LogsQueryService{gateway: gateway}
}

func (s *LogsQueryService) QueryLogs(ctx context.Context, req *v1.QueryLogsRequest) (*v1.QueryLogsResponse, error) {
	body, err := s.gateway.QueryLogs(ctx, req.GetDatasourceUid(), req.GetQuery(), req.GetStart(), req.GetEnd(), req.GetLimit(), req.GetDirection())
	if err != nil {
		return nil, err
	}
	return &v1.QueryLogsResponse{Body: body}, nil
}
