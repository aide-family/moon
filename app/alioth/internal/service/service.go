package service

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewHealthService,
	NewCaptchaService,
	NewAuthService,
	NewSelfService,
	NewNamespaceService,
	NewMemberService,
	NewUserService,
	NewRoleService,
	NewDatasourceService,
	NewMetricQueryService,
	NewAlertRuleService,
	NewAlertDictionaryService,
	NewAlertPageService,
	NewAlertEventService,
	NewEvalService,
	NewNotificationChannelService,
	NewNotificationTemplateService,
	NewNotificationGroupService,
	NewNotificationSubscriptionService,
	NewMessageLogService,
	NewNotifyService,
	NewAgentService,
	NewSSHCommandService,
	NewSFTPService,
	NewK8sProxyService,
)
