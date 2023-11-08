package conf

import (
	"github.com/google/wire"
)

// ProviderSetConf is conf providers.
var ProviderSetConf = wire.NewSet(
	wire.FieldsOf(new(*Bootstrap), "Server"),
	wire.FieldsOf(new(*Bootstrap), "Data"),
	wire.FieldsOf(new(*Bootstrap), "Env"),
	wire.FieldsOf(new(*Bootstrap), "Log"),
)
