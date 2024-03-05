package systemservice

import (
	"github.com/google/wire"
)

// ProviderSetSystem 注入系统相关服务
var ProviderSetSystem = wire.NewSet(
	NewUserService,
	NewRoleService,
	NewApiService,
	NewSyslogService,
	NewDictService,
)
