package convert

import (
	"encoding/json"

	"github.com/aide-family/magicbox/enum"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/data/impl/do"
)

func ToMachineInfoDO(in *bo.MachineInfoBo) (*do.MachineInfo, error) {
	if in == nil {
		return nil, nil
	}

	src := in.Source
	if src == enum.MachineInfoSource_MachineInfoSource_UNKNOWN {
		src = enum.MachineInfoSource_MachineInfoSource_ORIGIN
	}

	infoBytes, err := json.Marshal(in)
	if err != nil {
		return nil, err
	}

	return &do.MachineInfo{
		ID:          in.ID,
		MachineUUID: in.MachineUUID,
		HostName:    in.HostName,
		Source:      src,
		Info:        string(infoBytes),
	}, nil
}

func ToMachineInfoItemBo(row *do.MachineInfo) (*bo.MachineInfoBo, error) {
	if row == nil {
		return nil, nil
	}

	out := &bo.MachineInfoBo{}
	if row.Info != "" {
		if err := json.Unmarshal([]byte(row.Info), out); err != nil {
			return nil, err
		}
	}

	// Ensure dedup/filter keys are always populated from columns.
	if out.MachineUUID == "" {
		out.MachineUUID = row.MachineUUID
	}
	if out.HostName == "" {
		out.HostName = row.HostName
	}
	out.Source = row.Source
	return out, nil
}
