package impl

import (
	"github.com/google/wire"
)

// ProviderSetImpl is a provider set.
var ProviderSetImpl = wire.NewSet(
	NewCacheRepo,
	NewPingRepo,
	NewMetricRegister,
	NewScriptImpl,
)
