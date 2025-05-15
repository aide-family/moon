package service

import (
	"github.com/google/wire"
)

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewHealthService,
	NewSyncService,
	NewAlertService,
	NewEventBusService,
	NewLoadService,
	NewQueryService,
)
