package repository

import (
	"context"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/pkg/machine"
)

// MachineInfoProvider collects local machine info and persists/queries cluster machine info.
type MachineInfoProvider interface {
	// GetLocalMachineUUID returns the stable local machine UUID without collecting full machine info.
	GetLocalMachineUUID() string

	// Collect collects the local machine info.
	Collect(ctx context.Context) (*machine.MachineInfo, error)

	// GetMachineInfosByMachineUUIDs fetches existing machines by machine UUID.
	// It is used to implement merge semantics on reported payloads.
	GetMachineInfosByMachineUUIDs(ctx context.Context, machineUUIDs []string) ([]*machine.MachineInfo, error)

	// GetMachineInfoByMachineUUID fetches a machine by machine UUID.
	GetMachineInfoByMachineUUID(ctx context.Context, machineUUID string) (*machine.MachineInfo, error)

	// UpsertMachineInfos persists (insert or update) machines into storage.
	UpsertMachineInfos(ctx context.Context, machines []*machine.MachineInfo) error

	// UpdateLocalMachineInfo persists the local machine info.
	UpdateLocalMachineInfo(ctx context.Context, machine *machine.MachineInfo) error

	// ListMachineInfos returns a paginated view of machines in storage.
	ListMachineInfos(ctx context.Context, req *bo.ListMachineInfosBo) (*bo.PageResponseBo[*machine.MachineInfo], error)
}
