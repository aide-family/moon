package promservice

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/prom/notify"
)

type NotifyService struct {
	pb.UnimplementedNotifyServer

	log *log.Helper
}

func NewNotifyService(logger log.Logger) *NotifyService {
	return &NotifyService{
		log: log.NewHelper(log.With(logger, "module", "service.prom.notify")),
	}
}

func (s *NotifyService) CreateNotify(ctx context.Context, req *pb.CreateNotifyRequest) (*pb.CreateNotifyReply, error) {
	return &pb.CreateNotifyReply{}, nil
}

func (s *NotifyService) UpdateNotify(ctx context.Context, req *pb.UpdateNotifyRequest) (*pb.UpdateNotifyReply, error) {
	return &pb.UpdateNotifyReply{}, nil
}

func (s *NotifyService) DeleteNotify(ctx context.Context, req *pb.DeleteNotifyRequest) (*pb.DeleteNotifyReply, error) {
	return &pb.DeleteNotifyReply{}, nil
}

func (s *NotifyService) GetNotify(ctx context.Context, req *pb.GetNotifyRequest) (*pb.GetNotifyReply, error) {
	return &pb.GetNotifyReply{}, nil
}

func (s *NotifyService) ListNotify(ctx context.Context, req *pb.ListNotifyRequest) (*pb.ListNotifyReply, error) {
	return &pb.ListNotifyReply{}, nil
}

func (s *NotifyService) SelectNotify(ctx context.Context, req *pb.SelectNotifyRequest) (*pb.SelectNotifyReply, error) {
	return &pb.SelectNotifyReply{}, nil
}
