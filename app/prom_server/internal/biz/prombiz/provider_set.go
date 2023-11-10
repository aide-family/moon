package prombiz

import (
	"github.com/google/wire"
)

// ProviderSetPromBiz 注入prom biz
var ProviderSetPromBiz = wire.NewSet(
	NewStrategyBiz,
)
