package dictservice

import (
	"github.com/google/wire"
)

// ProviderSetDictService 注入DictService
var ProviderSetDictService = wire.NewSet(NewDictService)
