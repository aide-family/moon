package repository

import (
	"context"

	"github.com/aide-family/jade_tree/internal/biz/bo"
)

// MachineInfoProvider collects host runtime machine information.
type MachineInfoProvider interface {
	Collect(ctx context.Context) (*bo.MachineInfoBo, error)
}
