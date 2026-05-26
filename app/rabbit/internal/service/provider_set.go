// Package service is a service package for kratos.
package service

import "github.com/google/wire"

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewAlertService,
	NewHealthService,
	NewNamespaceService,
	NewSelfService,
	NewUserService,
	NewMemberService,
	NewCaptchaService,
	NewAuthService,
	NewEmailService,
	NewWebhookService,
	NewSenderService,
	NewMessageLogService,
	NewTemplateService,
	NewJobService,
	NewRecipientGroupService,
)
