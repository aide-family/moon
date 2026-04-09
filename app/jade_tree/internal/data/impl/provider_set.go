package impl

import "github.com/google/wire"

var ProviderSetImpl = wire.NewSet(
	NewHealthRepository,
	NewSSHRepository,
	NewAgentCommandDispatcher,
	NewSSHCommandRepository,
	NewCommandAuditRepository,
	NewMachineInfoRepository,
	NewProbeTaskRepository,
)
