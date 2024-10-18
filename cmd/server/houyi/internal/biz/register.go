package biz

import "github.com/google/wire"

// ProviderSetBiz is biz providers.
var ProviderSetBiz = wire.NewSet(
	NewGreeterUsecase,
	NewMetricBiz,
	NewStrategyBiz,
	NewAlertBiz,
	NewHeartbeatBiz,
)
