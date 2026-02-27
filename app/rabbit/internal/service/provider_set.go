// Package service is a service package for kratos.
package service

import "github.com/google/wire"

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewHealthService,
	NewNamespaceService,
	NewMemberService,
	NewAuthService,
	NewEmailService,
	NewWebhookService,
	NewSenderService,
	NewMessageLogService,
	NewTemplateService,
	NewJobService,
	NewRecipientGroupService,
)
