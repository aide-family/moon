package service

import (
	"github.com/google/wire"

	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/authorization"
	"github.com/aide-cloud/moon/cmd/server/palace/internal/service/user"
)

// ProviderSetService is service providers.
var ProviderSetService = wire.NewSet(
	NewGreeterService,
	user.NewUserService,
	authorization.NewAuthorizationService,
)
