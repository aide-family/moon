package repository

import (
	"context"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/pkg/machine"
)

// MachineInfoProvider collects local machine info and persists/queries cluster machine info.
type MachineInfoProvider interface {
	// GetLocalMachineIdentity returns machine UUID, hostname, and local IPv4 without a full hardware scan.
	GetLocalMachineIdentity() *bo.MachineInfoIdentityBo
	// GetLocalAgentInfo returns collector endpoint and version for the local machine.
	GetLocalAgentInfo(localIP string) *machine.MachineAgent

	// Collect collects the local machine info.
	Collect(ctx context.Context) (*machine.MachineInfo, error)

	// GetMachineInfoByIdentity fetches a machine by natural key (machine UUID + hostname + local IP).
	GetMachineInfoByIdentity(ctx context.Context, id *bo.MachineInfoIdentityBo) (*machine.MachineInfo, error)

	// UpsertMachineInfos persists (insert or update) machines into storage.
	UpsertMachineInfos(ctx context.Context, machines []*machine.MachineInfo) error

	// UpdateLocalMachineInfo persists the local machine info.
	UpdateLocalMachineInfo(ctx context.Context, machine *machine.MachineInfo) error

	// ListMachineInfos returns a paginated view of machines in storage.
	ListMachineInfos(ctx context.Context, req *bo.ListMachineInfosBo) (*bo.PageResponseBo[*machine.MachineInfo], error)
	// CountDispatchTargets returns the machine count matched by dispatch filter.
	CountDispatchTargets(ctx context.Context, filter *bo.DispatchSSHCommandFilterBo) (int64, error)
	// ListDispatchTargets returns matched machine rows for dispatch execution.
	ListDispatchTargets(ctx context.Context, filter *bo.DispatchSSHCommandFilterBo) ([]*machine.MachineInfo, error)
}
