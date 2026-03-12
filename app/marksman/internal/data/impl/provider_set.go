// Package impl is the implementation package for the Rabbit service.
package impl

import "github.com/google/wire"

// ProviderSetImpl is a set of providers.
var ProviderSetImpl = wire.NewSet(
	NewHealthRepository,
	NewNamespaceRepository,
	NewSelfRepository,
	NewUserRepository,
	NewMemberRepository,
	NewCaptchaRepository,
	NewMetricDatasourceQuerier,
	NewLevelRepository,
	NewDatasourceRepository,
	NewLoginRepository,
	NewStrategyGroupRepository,
	NewStrategyRepository,
	NewStrategyMetricRepository,
	NewJobChannel,
)
