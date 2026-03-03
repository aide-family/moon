// Package service is a service package for kratos.
package service

import "github.com/google/wire"

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewHealthService,
	NewAuthService,
	NewCaptchaService,
	NewNamespaceService,
	NewUserService,
	NewMemberService,
	NewSelfService,
)
