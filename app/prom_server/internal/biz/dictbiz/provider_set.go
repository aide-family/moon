package dictbiz

import (
	"github.com/google/wire"
)

// ProviderSetDictBiz 注入DictBiz
var ProviderSetDictBiz = wire.NewSet(NewBiz)
