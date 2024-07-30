package repoimpl

import (
	"github.com/google/wire"
)

// ProviderSetRepoImpl wire.ProviderSet
var ProviderSetRepoImpl = wire.NewSet(
	NewCacheRepo,
	NewHelloRepository,
)
