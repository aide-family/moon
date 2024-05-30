package service

import (
	"context"

	pb "github.com/aide-cloud/moon/api/houyi/metadata"
)

type MetricService struct {
	pb.UnimplementedMetricServer
}

func NewMetricService() *MetricService {
	return &MetricService{}
}

func (s *MetricService) Sync(ctx context.Context, req *pb.SyncRequest) (*pb.SyncReply, error) {
	return &pb.SyncReply{}, nil
}
