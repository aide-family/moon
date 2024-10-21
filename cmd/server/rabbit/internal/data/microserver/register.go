package microserver

import (
	"github.com/google/wire"
)

// ProviderSetRPCConn wire set
var ProviderSetRPCConn = wire.NewSet(
	NewPalaceConn,
)
