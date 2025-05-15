package rpc

import "github.com/google/wire"

// ProviderSetRpc is a provider set for rpc.
var ProviderSetRpc = wire.NewSet(
	NewServerRegisterRepo,
	NewCallbackRepo,
)
