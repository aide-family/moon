// Package impl is the implementation package for the Rabbit service.
package impl

import "github.com/google/wire"

// ProviderSetImpl is a set of providers.
var ProviderSetImpl = wire.NewSet(
	NewTransactionRepository,
	NewHealthRepository,
	NewNamespaceRepository,
	NewSelfRepository,
	NewUserRepository,
	NewMemberRepository,
	NewCaptchaRepository,
	NewMetricDatasourceQuerierRepository,
	NewMetricDatasourceProxyRepository,
	NewLevelRepository,
	NewDatasourceRepository,
	NewLoginRepository,
	NewStrategyGroupRepository,
	NewStrategyRepository,
	NewStrategyMetricRepository,
	NewEvaluateJobChannelRepository,
	NewAlertEventChannelRepository,
	NewAlertingEventChannelRepository,
	NewAlertPageRepository,
	NewAlertEventRepository,
	NewUserAlertPageRepository,
	NewNotificationGroupRepository,
	NewNotificationGroupSubscriptionRepository,
	NewRabbitWebhookRepository,
	NewRabbitTemplateRepository,
	NewRabbitSenderRepository,
)
