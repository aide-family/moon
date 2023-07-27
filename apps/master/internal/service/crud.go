package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/strategy/v1"
)

type (
	ICrudLogic interface {
		CreateCrud(ctx context.Context, req *pb.CreateCrudRequest) (*pb.CreateCrudReply, error)
		UpdateCrud(ctx context.Context, req *pb.UpdateCrudRequest) (*pb.UpdateCrudReply, error)
		DeleteCrud(ctx context.Context, req *pb.DeleteCrudRequest) (*pb.DeleteCrudReply, error)
		GetCrud(ctx context.Context, req *pb.GetCrudRequest) (*pb.GetCrudReply, error)
	}

	CrudService struct {
		pb.UnimplementedCrudServer

		logger *log.Helper
		logic  ICrudLogic
	}
)

var _ pb.CrudServer = (*CrudService)(nil)

func NewCrudService(logic ICrudLogic, logger log.Logger) *CrudService {
	return &CrudService{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Crud"))}
}

func (l *CrudService) CreateCrud(ctx context.Context, req *pb.CreateCrudRequest) (*pb.CreateCrudReply, error) {
	l.logger.Debugf("CreateCrud req: %v", req)
	return l.logic.CreateCrud(ctx, req)
}

func (l *CrudService) UpdateCrud(ctx context.Context, req *pb.UpdateCrudRequest) (*pb.UpdateCrudReply, error) {
	l.logger.Debugf("UpdateCrud req: %v", req)
	return l.logic.UpdateCrud(ctx, req)
}

func (l *CrudService) DeleteCrud(ctx context.Context, req *pb.DeleteCrudRequest) (*pb.DeleteCrudReply, error) {
	l.logger.Debugf("DeleteCrud req: %v", req)
	return l.logic.DeleteCrud(ctx, req)
}

func (l *CrudService) GetCrud(ctx context.Context, req *pb.GetCrudRequest) (*pb.GetCrudReply, error) {
	l.logger.Debugf("GetCrud req: %v", req)
	return l.logic.GetCrud(ctx, req)
}
