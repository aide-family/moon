package biz

import (
	"context"

	"github.com/google/wire"

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
