package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	pb "prometheus-manager/api/strategy/v1"
	"prometheus-manager/apps/master/internal/service"
)

type (
	ICrudRepo interface {
		V1Repo
	}

	CrudLogic struct {
		logger *log.Helper
		repo   ICrudRepo
	}
)

var _ service.ICrudLogic = (*CrudLogic)(nil)

func NewCrudLogic(repo ICrudRepo, logger log.Logger) *CrudLogic {
	return &CrudLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/Crud"))}
}

func (s *CrudLogic) CreateCrud(ctx context.Context, req *pb.CreateCrudRequest) (*pb.CreateCrudReply, error) {
	return nil, nil
}

func (s *CrudLogic) UpdateCrud(ctx context.Context, req *pb.UpdateCrudRequest) (*pb.UpdateCrudReply, error) {
	return nil, nil
}

func (s *CrudLogic) DeleteCrud(ctx context.Context, req *pb.DeleteCrudRequest) (*pb.DeleteCrudReply, error) {
	return nil, nil
}

func (s *CrudLogic) GetCrud(ctx context.Context, req *pb.GetCrudRequest) (*pb.GetCrudReply, error) {
	return nil, nil
}
