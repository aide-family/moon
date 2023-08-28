package biz

import (
	"context"
	"prometheus-manager/apps/node/internal/conf"
	"prometheus-manager/pkg/util/dir"

	"github.com/google/wire"
	pb "prometheus-manager/api/alert/v1"

	"prometheus-manager/pkg/alert"
	"prometheus-manager/pkg/times"

	"prometheus-manager/apps/node/internal/service"
)

type V1Repo interface {
	V1(ctx context.Context) string
}

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewPushLogic,
	wire.Bind(new(service.IPushLogic), new(*PushLogic)),
	NewPullLogic,
	wire.Bind(new(service.IPullLogic), new(*PullLogic)),
	NewLoadLogic,
	wire.Bind(new(service.ILoadLogic), new(*LoadLogic)),
	NewPingLogic,
	wire.Bind(new(service.IPingLogic), new(*PingLogic)),
	NewAlertLogic,
	wire.Bind(new(service.IAlertLogic), new(*AlertLogic)),
)

const (
	loadModuleName  = "biz/load"
	pingModuleName  = "biz/ping"
	pullModuleName  = "biz/pull"
	pushModuleName  = "biz/push"
	alertModuleName = "biz/alert"
)

func alertWebhookRequestToAlertData(req *pb.WebhookRequest) *alert.Data {
	alerts := make([]*alert.Alert, 0, len(req.GetAlerts()))
	for _, info := range req.GetAlerts() {
		alerts = append(alerts, &alert.Alert{
			Status:       info.GetStatus(),
			Labels:       info.GetLabels(),
			Annotations:  info.GetAnnotations(),
			StartsAt:     times.TimeToUnix(alert.ParseTime(info.GetStartsAt())),
			EndsAt:       times.TimeToUnix(alert.ParseTime(info.GetEndsAt())),
			GeneratorURL: info.GetGeneratorURL(),
			Fingerprint:  info.GetFingerprint(),
		})
	}

	return &alert.Data{
		Receiver:          req.GetReceiver(),
		Status:            req.GetStatus(),
		Alerts:            alerts,
		GroupLabels:       req.GetGroupLabels(),
		CommonLabels:      req.GetCommonLabels(),
		CommonAnnotations: req.GetCommonAnnotations(),
		ExternalURL:       req.GetExternalURL(),
		Version:           req.GetVersion(),
		GroupKey:          req.GetGroupKey(),
		TruncatedAlerts:   req.GetTruncatedAlerts(),
	}
}

func joinUniKey(nodeName, path string) string {
	return nodeName + "__" + path
}

func getStrategyPathMap(datasource []*conf.PromDatasource) map[string]struct{} {
	strategyPathMap := make(map[string]struct{})
	for _, promDatasource := range datasource {
		strategyPath := promDatasource.GetPath()
		if len(strategyPath) == 0 {
			continue
		}
		for _, p := range strategyPath {
			strategyPathMap[joinUniKey(promDatasource.GetName(), dir.RemoveLastSlash(p))] = struct{}{}
		}
	}

	return strategyPathMap
}
