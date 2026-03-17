// Package service is a service package for kratos.
package service

import "github.com/google/wire"

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewHealthService,
	NewNamespaceService,
	NewSelfService,
	NewUserService,
	NewMemberService,
	NewCaptchaService,
	NewLevelService,
	NewDatasourceService,
	NewAuthService,
	NewStrategyService,
	NewStrategyMetricService,
	NewEvaluateService,
	NewAlertService,
)
