package biz

import (
	"github.com/google/wire"
	"prometheus-manager/apps/node/internal/service"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewPushLogic,
	wire.Bind(new(service.IPushLogic), new(*PushLogic)),
	NewPullLogic,
	wire.Bind(new(service.IPullLogic), new(*PullLogic)),
	NewLoadLogic,
	wire.Bind(new(service.ILoadLogic), new(*LoadLogic)),
)
