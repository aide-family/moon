// Package biz is the business logic for the rabbit service.
package biz

import "github.com/google/wire"

var ProviderSetBiz = wire.NewSet(
	NewHealth,
	NewNamespace,
	NewMember,
	NewLoginBiz,
	NewJob,
	NewEmailConfig,
	NewEmail,
	NewWebhookConfig,
	NewWebhook,
	NewMessageLog,
	NewMessage,
	NewTemplate,
	NewRecipientGroup,
)
