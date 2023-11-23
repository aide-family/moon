package biz

import (
	"github.com/google/wire"
)

// ProviderSetBiz 注入biz依赖
var ProviderSetBiz = wire.NewSet(
	NewPingRepo,
	NewDictBiz,
	NewHistoryBiz,
	NewPageBiz,
	NewStrategyBiz,
	NewStrategyGroupBiz,
	NewUserBiz,
	NewRoleBiz,
)
