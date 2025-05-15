package rpc

import "github.com/google/wire"

// ProviderSetRPC is a set of RPC providers.
var ProviderSetRPC = wire.NewSet(
	NewRabbitServer,
	NewHouyiServer,
)
