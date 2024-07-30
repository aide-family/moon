package microserverrepoimpl

import (
	"github.com/google/wire"
)

// ProviderSetRPCRepoImpl rpc repo impl provider
var ProviderSetRPCRepoImpl = wire.NewSet(
	NewDatasourceMetricRepository,
	NewMsgRepository,
)
