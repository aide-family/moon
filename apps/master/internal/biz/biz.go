package biz

import (
	"context"
	"github.com/google/wire"
	"prometheus-manager/apps/master/internal/service"
)

// ProviderSet is biz providers.
var ProviderSet = wire.NewSet(
	NewCrudLogic,
	wire.Bind(new(service.ICrudLogic), new(*CrudLogic)),
	NewPingLogic,
	wire.Bind(new(service.IPingLogic), new(*PingLogic)),
)

type V1Repo interface {
	V1(ctx context.Context) string
}
