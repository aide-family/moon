package impl

import "github.com/google/wire"

var ProviderSetImpl = wire.NewSet(
	NewHealthRepository,
	NewSSHRepository,
	NewSSHCommandRepository,
	NewCommandAuditRepository,
	NewMachineInfoRepository,
	NewProbeTaskRepository,
)
