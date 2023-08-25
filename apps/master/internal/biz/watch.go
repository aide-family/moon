package biz

import (
	"context"
	"prometheus-manager/api"
	"prometheus-manager/pkg/alert"

	"github.com/go-kratos/kratos/v2/log"
	"go.opentelemetry.io/otel"

	pb "prometheus-manager/api/alert/v1"

	"prometheus-manager/apps/master/internal/service"
)

type (
	IWatchRepo interface {
		V1Repo

		WatchAlert(ctx context.Context, req *alert.Data) error
	}

	WatchLogic struct {
		logger *log.Helper
		repo   IWatchRepo
	}
)

var _ service.IWatchLogic = (*WatchLogic)(nil)

func NewWatchLogic(repo IWatchRepo, logger log.Logger) *WatchLogic {
	return &WatchLogic{repo: repo, logger: log.NewHelper(log.With(logger, "module", watchModuleName))}
}

func (s *WatchLogic) WatchAlert(ctx context.Context, req *pb.WatchRequest) (*pb.WatchReply, error) {
	ctx, span := otel.Tracer(watchModuleName).Start(ctx, "WatchLogic.WatchAlert")
	defer span.End()

	// TODO 落库、落ES、落Redis、发送通知
	if err := s.repo.WatchAlert(ctx, watchRequestToData(req)); err != nil {
		s.logger.Errorf("WatchAlert error: %v", err)
		return nil, err
	}

	return &pb.WatchReply{Response: &api.Response{Message: "succeed"}}, nil
}

func watchRequestToData(req *pb.WatchRequest) *alert.Data {
	if req == nil {
		return nil
	}

	alertList := make([]*alert.Alert, 0, len(req.GetAlerts()))
	for _, alertInfo := range req.GetAlerts() {
		alertList = append(alertList, &alert.Alert{
			Status:       alertInfo.GetStatus(),
			Labels:       alertInfo.GetLabels(),
			Annotations:  alertInfo.GetAnnotations(),
			StartsAt:     alertInfo.GetStartsAt(),
			EndsAt:       alertInfo.GetEndsAt(),
			GeneratorURL: alertInfo.GetGeneratorURL(),
			Fingerprint:  alertInfo.GetFingerprint(),
		})
	}

	return &alert.Data{
		Receiver:          req.GetReceiver(),
		Status:            req.GetStatus(),
		Alerts:            alertList,
		GroupLabels:       req.GetGroupLabels(),
		CommonLabels:      req.GetCommonLabels(),
		CommonAnnotations: req.GetCommonAnnotations(),
		ExternalURL:       req.GetExternalURL(),
		Version:           req.GetVersion(),
		GroupKey:          req.GetGroupKey(),
		TruncatedAlerts:   req.GetTruncatedAlerts(),
	}
}
