package service

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
)

type (
	IAlarmPageV1Logic interface {
		CreateAlarmPage(ctx context.Context, req *pb.CreateAlarmPageRequest) (*pb.CreateAlarmPageReply, error)
		UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error)
		UpdateAlarmPagesStatus(ctx context.Context, req *pb.UpdateAlarmPagesStatusRequest) (*pb.UpdateAlarmPagesStatusReply, error)
		DeleteAlarmPage(ctx context.Context, req *pb.DeleteAlarmPageRequest) (*pb.DeleteAlarmPageReply, error)
		GetAlarmPage(ctx context.Context, req *pb.GetAlarmPageRequest) (*pb.GetAlarmPageReply, error)
		ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error)
	}

	AlarmPageV1Service struct {
		pb.UnimplementedAlarmPageServer

		logger *log.Helper
		logic  IAlarmPageV1Logic
	}
)

var _ pb.AlarmPageServer = (*AlarmPageV1Service)(nil)

func NewAlarmPageService(logic IAlarmPageV1Logic, logger log.Logger) *AlarmPageV1Service {
	return &AlarmPageV1Service{logic: logic, logger: log.NewHelper(log.With(logger, "module", "service/AlarmPage"))}
}

func (l *AlarmPageV1Service) CreateAlarmPage(ctx context.Context, req *pb.CreateAlarmPageRequest) (*pb.CreateAlarmPageReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "AlarmPageV1Service.CreateAlarmPage")
	defer span.End()
	return l.logic.CreateAlarmPage(ctx, req)
}

func (l *AlarmPageV1Service) UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "AlarmPageV1Service.UpdateAlarmPage")
	defer span.End()
	return l.logic.UpdateAlarmPage(ctx, req)
}

func (l *AlarmPageV1Service) UpdateAlarmPagesStatus(ctx context.Context, req *pb.UpdateAlarmPagesStatusRequest) (*pb.UpdateAlarmPagesStatusReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "AlarmPageV1Service.UpdateAlarmPagesStatus")
	defer span.End()
	return l.logic.UpdateAlarmPagesStatus(ctx, req)
}

func (l *AlarmPageV1Service) DeleteAlarmPage(ctx context.Context, req *pb.DeleteAlarmPageRequest) (*pb.DeleteAlarmPageReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "AlarmPageV1Service.DeleteAlarmPage")
	defer span.End()
	return l.logic.DeleteAlarmPage(ctx, req)
}

func (l *AlarmPageV1Service) GetAlarmPage(ctx context.Context, req *pb.GetAlarmPageRequest) (*pb.GetAlarmPageReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "AlarmPageV1Service.GetAlarmPage")
	defer span.End()
	return l.logic.GetAlarmPage(ctx, req)
}

func (l *AlarmPageV1Service) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	ctx, span := otel.Tracer("service").Start(ctx, "AlarmPageV1Service.ListAlarmPage")
	defer span.End()
	return l.logic.ListAlarmPage(ctx, req)
}
