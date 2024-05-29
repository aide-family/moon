package microserverrepoimpl

import (
	"github.com/google/wire"
)

var ProviderSetRpcRepoImpl = wire.NewSet(
	NewDatasourceMetricRepository,
	NewMsgRepository,
)
