package service

import "github.com/google/wire"

var ProviderSetService = wire.NewSet(NewHealthService)
