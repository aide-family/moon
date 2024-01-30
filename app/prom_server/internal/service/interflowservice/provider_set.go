package interflowservice

import (
	"github.com/google/wire"
)

// ProviderSetInterflowService 注入InterflowService
var ProviderSetInterflowService = wire.NewSet(NewHookInterflowService)
