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
)
