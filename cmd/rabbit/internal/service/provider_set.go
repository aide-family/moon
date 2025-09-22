// Package service is a service package for kratos.
package service

import "github.com/google/wire"

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewHealthService,
	NewSendService,
	NewSyncService,
	NewAlertService,
)
