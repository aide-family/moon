package biz

import "github.com/google/wire"

// ProviderSetBiz is biz providers.
var ProviderSetBiz = wire.NewSet(
	NewMetricBiz,
	NewStrategyBiz,
	NewAlertBiz,
	NewHeartbeatBiz,
)
