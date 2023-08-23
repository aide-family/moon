package biz

import (
	"context"
	"encoding/json"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api/alert/v1"

	"prometheus-manager/apps/node/internal/service"
)

type (
	IAlertRepo interface {
		V1Repo
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

	str, _ := json.Marshal(req)
	s.logger.Info("Webhook", string(str))

	return nil, nil
}
