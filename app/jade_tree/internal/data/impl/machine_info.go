// Package impl provides repository implementations.
package impl

import (
	"context"
	"strings"

	"github.com/aide-family/magicbox/merr"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/conf"
	"github.com/aide-family/jade_tree/internal/data"
	"github.com/aide-family/jade_tree/internal/data/impl/query"
	"github.com/aide-family/jade_tree/pkg/machine"
)

func NewMachineInfoRepository(bc *conf.Bootstrap, d *data.Data) repository.MachineInfoProvider {
	query.SetDefault(d.DB())
	var enabledCollectSelf bool
	if bc != nil && bc.GetCollectSelf() != nil {
		enabledCollectSelf = strings.EqualFold(bc.GetCollectSelf().GetEnabled(), "true")
	}
	return &machineInfoRepository{Data: d, enabledCollectSelf: enabledCollectSelf}
}

type machineInfoRepository struct {
	*data.Data
	enabledCollectSelf bool
}

func (m *machineInfoRepository) Collect(ctx context.Context) (*machine.MachineInfo, error) {
	if !m.enabledCollectSelf {
		return nil, merr.ErrorParams("collect self is not enabled")
	}
	return machine.Collect(ctx)
}

func (m *machineInfoRepository) GetLocalMachineIdentity() *bo.MachineInfoIdentityBo {
	u, h, lip := machine.LocalMachineIdentity()
	return &bo.MachineInfoIdentityBo{MachineUUID: u, HostName: h, LocalIP: lip}
}
