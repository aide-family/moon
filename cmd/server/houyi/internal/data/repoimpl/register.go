package repoimpl

import (
	"github.com/google/wire"
)

// ProviderSetRepoImpl wire Set
var ProviderSetRepoImpl = wire.NewSet(
	NewCacheRepo,
	NewMetricRepository,
	NewStrategyRepository,
	NewAlertRepository,
)
