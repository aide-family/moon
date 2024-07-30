package repoimpl

import (
	"github.com/google/wire"
)

// ProviderSetRepoImpl wire set
var ProviderSetRepoImpl = wire.NewSet(
	NewCacheRepo,
)
