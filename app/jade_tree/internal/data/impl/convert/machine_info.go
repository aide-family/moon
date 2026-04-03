package convert

import (
	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/jade_tree/internal/data/impl/do"
	"github.com/aide-family/jade_tree/pkg/machine"
)

func ToMachineInfoDO(in *machine.MachineInfo) (*do.MachineInfo, error) {
	if in == nil {
		return nil, nil
	}

	src := in.Source
	if src == enum.MachineInfoSource_MachineInfoSource_UNKNOWN {
		src = enum.MachineInfoSource_MachineInfoSource_ORIGIN
	}

	localIP := ""
	if in.Network != nil {
		localIP = in.Network.LocalIP
	}

	return &do.MachineInfo{
		ID:          in.ID,
		MachineUUID: in.MachineUUID,
		HostName:    in.HostName,
		LocalIP:     localIP,
		Source:      src,
		Info:        in,
	}, nil
}

func ToMachineInfoItemBo(row *do.MachineInfo) (*machine.MachineInfo, error) {
	if row == nil {
		return nil, nil
	}

	return row.Info, nil
}
