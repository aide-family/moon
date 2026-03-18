// Package impl is the implementation package for the Rabbit service.
package impl

import "github.com/google/wire"

// ProviderSetImpl is a set of providers.
var ProviderSetImpl = wire.NewSet(
	NewTransaction,
	NewHealthRepository,
	NewNamespaceRepository,
	NewSelfRepository,
	NewUserRepository,
	NewMemberRepository,
	NewCaptchaRepository,
	NewMetricDatasourceQuerier,
	NewMetricDatasourceProxy,
	NewLevelRepository,
	NewDatasourceRepository,
	NewLoginRepository,
	NewStrategyGroupRepository,
	NewStrategyRepository,
	NewStrategyMetricRepository,
	NewJobChannel,
	NewAlertEventChannel,
	NewAlertingRepository,
	NewAlertPageRepository,
	NewAlertEventRepository,
	NewUserAlertPageRepository,
	NewNotificationGroupRepository,
	NewRabbitWebhookRepository,
	NewRabbitTemplateRepository,
	NewRabbitSenderRepository,
)
