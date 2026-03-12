// Package biz is the business logic for the marksman service.
package biz

import "github.com/google/wire"

var ProviderSetBiz = wire.NewSet(
	NewHealth,
	NewNamespace,
	NewSelf,
	NewUser,
	NewMember,
	NewCaptcha,
	NewLevel,
	NewDatasource,
	NewLoginBiz,
	NewStrategy,
	NewStrategyMetric,
	NewEvaluateBiz,
)
