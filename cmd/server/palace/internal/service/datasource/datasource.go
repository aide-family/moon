package datasource

import (
	"context"

	pb "github.com/aide-cloud/moon/api/admin/datasource"
)

type Service struct {
	pb.UnimplementedDatasourceServer
}

func NewDatasourceService() *Service {
	return &Service{}
}

func (s *Service) CreateDatasource(ctx context.Context, req *pb.CreateDatasourceRequest) (*pb.CreateDatasourceReply, error) {
	return &pb.CreateDatasourceReply{}, nil
}

func (s *Service) UpdateDatasource(ctx context.Context, req *pb.UpdateDatasourceRequest) (*pb.UpdateDatasourceReply, error) {
	return &pb.UpdateDatasourceReply{}, nil
}

func (s *Service) DeleteDatasource(ctx context.Context, req *pb.DeleteDatasourceRequest) (*pb.DeleteDatasourceReply, error) {
	return &pb.DeleteDatasourceReply{}, nil
}

func (s *Service) GetDatasource(ctx context.Context, req *pb.GetDatasourceRequest) (*pb.GetDatasourceReply, error) {
	return &pb.GetDatasourceReply{}, nil
}

func (s *Service) ListDatasource(ctx context.Context, req *pb.ListDatasourceRequest) (*pb.ListDatasourceReply, error) {
	return &pb.ListDatasourceReply{}, nil
}

func (s *Service) UpdateDatasourceStatus(ctx context.Context, req *pb.UpdateDatasourceStatusRequest) (*pb.UpdateDatasourceStatusReply, error) {
	return &pb.UpdateDatasourceStatusReply{}, nil
}

func (s *Service) GetDatasourceSelect(ctx context.Context, req *pb.GetDatasourceSelectRequest) (*pb.GetDatasourceSelectReply, error) {
	return &pb.GetDatasourceSelectReply{}, nil
}
