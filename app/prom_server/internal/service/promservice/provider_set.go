package promservice

import (
	"github.com/google/wire"
)

// ProviderSetProm 注入prometheus相关服务
var ProviderSetProm = wire.NewSet(
	NewStrategyService,
)
