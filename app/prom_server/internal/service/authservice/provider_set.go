package authservice

import (
	"github.com/google/wire"
)

// ProviderSetAuthService 注入AuthService
var ProviderSetAuthService = wire.NewSet(NewAuthService)
