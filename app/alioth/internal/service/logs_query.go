package service

import (
	"context"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/internal/biz/datasource"
)

type LogsQueryService struct {
	v1.UnimplementedLogsQueryServiceServer
	datasources *datasource.Usecase
}

func NewLogsQueryService(uc *datasource.Usecase) *LogsQueryService {
	return &LogsQueryService{datasources: uc}
}

func (s *LogsQueryService) QueryLogs(ctx context.Context, req *v1.QueryLogsRequest) (*v1.QueryLogsResponse, error) {
	body, err := s.datasources.QueryLogs(ctx, req.GetDatasourceUid(), req.GetQuery(), req.GetStart(), req.GetEnd(), req.GetLimit(), req.GetDirection())
	if err != nil {
		return nil, err
	}
	return &v1.QueryLogsResponse{Body: body}, nil
}
