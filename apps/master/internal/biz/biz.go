package biz

import (
	"context"

	"github.com/google/wire"

	"prometheus-manager/apps/master/internal/service"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewPingLogic,
	wire.Bind(new(service.IPingLogic), new(*PingLogic)),
	NewPushLogic,
	wire.Bind(new(service.IPushLogic), new(*PushLogic)),
	NewPromLogic,
	wire.Bind(new(service.IPromLogic), new(*PromLogic)),
)

type V1Repo interface {
	V1(ctx context.Context) string
}
