package portal_service

import "github.com/google/wire"

var ProviderSetPortalService = wire.NewSet(
	NewAuthService,
	NewHomeService,
	NewPricingService,
)
