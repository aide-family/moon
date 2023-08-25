package biz

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"
	"prometheus-manager/api"

	pb "prometheus-manager/api/alert/v1"
	"prometheus-manager/pkg/alert"

	"prometheus-manager/apps/node/internal/service"
)

type (
	IAlertRepo interface {
		V1Repo

		SyncAlert(ctx context.Context, alertData *alert.Data) error
	}

	AlertLogic struct {
		logger *log.Helper
		repo   IAlertRepo
	}
)

var _ service.IAlertLogic = (*AlertLogic)(nil)

func NewAlertLogic(repo IAlertRepo, logger log.Logger) *AlertLogic {
	return &AlertLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", alertModuleName))}
}

func (s *AlertLogic) Webhook(ctx context.Context, req *pb.WebhookRequest) (*pb.WebhookReply, error) {
	ctx, span := otel.Tracer(alertModuleName).Start(ctx, "AlertLogic.Webhook")
	defer span.End()

	if err := s.repo.SyncAlert(ctx, alertWebhookRequestToAlertData(req)); err != nil {
		s.logger.WithContext(ctx).Errorf("SyncAlert err: %v", err)
		return nil, err
	}

	return &pb.WebhookReply{Response: &api.Response{Message: "succeed"}}, nil
}
