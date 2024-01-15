package repositiryimpl

import (
	"github.com/google/wire"
)

var ProviderSetRepoImpl = wire.NewSet(NewAlarmRepo)
