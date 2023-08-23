package service

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api/alert/v1"
)

type (
	IAlertLogic interface {
		Webhook(ctx context.Context, req *pb.WebhookRequest) (*pb.WebhookReply, error)
	}

	AlertService struct {
		pb.UnimplementedAlertServer

		logger *log.Helper
		logic  IAlertLogic
	}
)

var _ pb.AlertServer = (*AlertService)(nil)

func NewAlertService(logic IAlertLogic, logger log.Logger) *AlertService {
	return &AlertService{logic: logic, logger: log.NewHelper(log.With(logger, "module", alertModuleName))}
}

func (l *AlertService) Webhook(ctx context.Context, req *pb.WebhookRequest) (*pb.WebhookReply, error) {
	ctx, span := otel.Tracer(alertModuleName).Start(ctx, "AlertService.Webhook")
	defer span.End()
	return l.logic.Webhook(ctx, req)
}
