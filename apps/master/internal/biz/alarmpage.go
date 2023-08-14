package biz

import (
	"context"
	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	pb "prometheus-manager/api/prom/v1"
	"prometheus-manager/apps/master/internal/service"
)

type (
	IAlarmPageV1Repo interface {
		V1Repo
	}

	AlarmPageLogic struct {
		logger *log.Helper
		repo   IAlarmPageV1Repo
	}
)

var _ service.IAlarmPageV1Logic = (*AlarmPageLogic)(nil)

func NewAlarmPageLogic(repo IAlarmPageV1Repo, logger log.Logger) *AlarmPageLogic {
	return &AlarmPageLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", "biz/AlarmPage"))}
}

func (s *AlarmPageLogic) CreateAlarmPage(ctx context.Context, req *pb.CreateAlarmPageRequest) (*pb.CreateAlarmPageReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "AlarmPageLogic.CreateAlarmPage")
	defer span.End()
	return nil, nil
}
func (s *AlarmPageLogic) UpdateAlarmPage(ctx context.Context, req *pb.UpdateAlarmPageRequest) (*pb.UpdateAlarmPageReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "AlarmPageLogic.UpdateAlarmPage")
	defer span.End()
	return nil, nil
}
func (s *AlarmPageLogic) DeleteAlarmPage(ctx context.Context, req *pb.DeleteAlarmPageRequest) (*pb.DeleteAlarmPageReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "AlarmPageLogic.DeleteAlarmPage")
	defer span.End()
	return nil, nil
}
func (s *AlarmPageLogic) GetAlarmPage(ctx context.Context, req *pb.GetAlarmPageRequest) (*pb.GetAlarmPageReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "AlarmPageLogic.GetAlarmPage")
	defer span.End()
	return nil, nil
}
func (s *AlarmPageLogic) ListAlarmPage(ctx context.Context, req *pb.ListAlarmPageRequest) (*pb.ListAlarmPageReply, error) {
	ctx, span := otel.Tracer("biz").Start(ctx, "AlarmPageLogic.ListAlarmPage")
	defer span.End()
	return nil, nil
}
