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
	osType := ""
	if in.System != nil {
		osType = in.System.OS
	}
	agentEndpoint := ""
	agentVersion := ""
	if in.Agent != nil {
		agentEndpoint = in.Agent.HTTPEndpoint
		if agentEndpoint == "" {
			agentEndpoint = in.Agent.Endpoint
		}
		agentVersion = in.Agent.Version
	}

	return &do.MachineInfo{
		ID:            in.ID,
		MachineUUID:   in.MachineUUID,
		HostName:      in.HostName,
		LocalIP:       localIP,
		Source:        src,
		OSType:        osType,
		AgentEndpoint: agentEndpoint,
		AgentVersion:  agentVersion,
		Info:          in,
	}, nil
}

func ToMachineInfoItemBo(row *do.MachineInfo) (*machine.MachineInfo, error) {
	if row == nil {
		return nil, nil
	}

	item := row.Info
	if item == nil {
		item = &machine.MachineInfo{}
	}
	item.ID = row.ID
	if item.Agent == nil {
		item.Agent = &machine.MachineAgent{}
	}
	if row.AgentEndpoint != "" {
		item.Agent.Endpoint = row.AgentEndpoint
		if item.Agent.HTTPEndpoint == "" {
			item.Agent.HTTPEndpoint = row.AgentEndpoint
		}
	}
	if row.AgentVersion != "" {
		item.Agent.Version = row.AgentVersion
	}
	if item.Agent.Endpoint == "" {
		switch {
		case item.Agent.HTTPEndpoint != "" && item.Agent.GRPCEndpoint != "":
			item.Agent.Endpoint = item.Agent.HTTPEndpoint + "," + item.Agent.GRPCEndpoint
		case item.Agent.HTTPEndpoint != "":
			item.Agent.Endpoint = item.Agent.HTTPEndpoint
		case item.Agent.GRPCEndpoint != "":
			item.Agent.Endpoint = item.Agent.GRPCEndpoint
		}
	}
	return item, nil
}
