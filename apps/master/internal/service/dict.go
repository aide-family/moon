package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	IDictV1Logic interface {
		CreateDict(ctx context.Context, req *pb.CreateDictRequest) (*pb.CreateDictReply, error)
		UpdateDict(ctx context.Context, req *pb.UpdateDictRequest) (*pb.UpdateDictReply, error)
		DeleteDict(ctx context.Context, req *pb.DeleteDictRequest) (*pb.DeleteDictReply, error)
		GetDict(ctx context.Context, req *pb.GetDictRequest) (*pb.GetDictReply, error)
		ListDict(ctx context.Context, req *pb.ListDictRequest) (*pb.ListDictReply, error)
	}

	DictV1Service struct {
		pb.UnimplementedDictServer

		logger *log.Helper
		logic  IDictV1Logic
	}
)

var _ pb.DictServer = (*DictV1Service)(nil)

func NewDictService(logic IDictV1Logic, logger log.Logger) *DictV1Service {
	return &DictV1Service{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/Dict"))}
}

func (l *DictV1Service) CreateDict(ctx context.Context, req *pb.CreateDictRequest) (*pb.CreateDictReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DictV1Service.CreateDict")
	defer span.End()
	return l.logic.CreateDict(ctx, req)
}

func (l *DictV1Service) UpdateDict(ctx context.Context, req *pb.UpdateDictRequest) (*pb.UpdateDictReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DictV1Service.UpdateDict")
	defer span.End()
	return l.logic.UpdateDict(ctx, req)
}

func (l *DictV1Service) DeleteDict(ctx context.Context, req *pb.DeleteDictRequest) (*pb.DeleteDictReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DictV1Service.DeleteDict")
	defer span.End()
	return l.logic.DeleteDict(ctx, req)
}

func (l *DictV1Service) GetDict(ctx context.Context, req *pb.GetDictRequest) (*pb.GetDictReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DictV1Service.GetDict")
	defer span.End()
	return l.logic.GetDict(ctx, req)
}

func (l *DictV1Service) ListDict(ctx context.Context, req *pb.ListDictRequest) (*pb.ListDictReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "DictV1Service.ListDict")
	defer span.End()
	return l.logic.ListDict(ctx, req)
}
