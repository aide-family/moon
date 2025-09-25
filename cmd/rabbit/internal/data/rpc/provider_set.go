// Package rpc is a rpc package for kratos.
package rpc

import "github.com/google/wire"

// ProviderSetRPC is a provider set for rpc.
var ProviderSetRPC = wire.NewSet(
	NewServerRegisterRepo,
)
