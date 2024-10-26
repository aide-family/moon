package service

import (
	"github.com/google/wire"
)

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewMetricService,
	NewHealthService,
	NewStrategyService,
	NewAlertService,
)
