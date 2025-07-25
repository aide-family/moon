package biz

import (
	"github.com/google/wire"
)

// ProviderSetBiz set biz dependency
var ProviderSetBiz = wire.NewSet(
	NewHealthBiz,
	NewRegisterBiz,
	NewMetricManagerBiz,
	NewScriptBiz,
)
