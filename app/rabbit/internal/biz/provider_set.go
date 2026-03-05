// Package biz is the business logic for the rabbit service.
package biz

import "github.com/google/wire"

var ProviderSetBiz = wire.NewSet(
	NewHealth,
	NewNamespace,
	NewSelf,
	NewUser,
	NewMember,
	NewCaptcha,
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
