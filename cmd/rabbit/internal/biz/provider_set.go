package biz

import (
	"github.com/google/wire"
)

// ProviderSetBiz set biz dependency
var ProviderSetBiz = wire.NewSet(
	NewHealthBiz,
	NewRegisterBiz,
	NewConfig,
	NewEmail,
	NewSMS,
	NewHook,
	NewLock,
	NewAlert,
)
