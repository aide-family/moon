package repository

import (
	"github.com/google/wire"
)

// ProviderSetRepository 注入repository依赖
var ProviderSetRepository = wire.NewSet(
	NewDictRepo,
	NewStrategyRepo,
	NewAlarmPageRepo,
	NewAlarmHistoryRepo,
)
