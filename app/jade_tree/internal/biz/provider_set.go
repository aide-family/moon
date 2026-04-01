package biz

import "github.com/google/wire"

var ProviderSetBiz = wire.NewSet(NewHealth, NewSSHCommand)
