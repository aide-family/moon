package alarmbiz

import (
	"github.com/google/wire"
)

// ProviderSetAlarmBiz 注入DictBiz
var ProviderSetAlarmBiz = wire.NewSet(
	NewPageBiz,
	NewHistoryBiz,
)
