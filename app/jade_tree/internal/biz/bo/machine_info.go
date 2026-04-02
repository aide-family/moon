package bo

import apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"

type MachineCPUCoreBo struct {
	ID              int32
	HardwareThreads int32
	LogicalCPU      []int32
}

type MachineCPUProcessorBo struct {
	ID                  int32
	Vendor              string
	Model               string
	TotalCores          int32
	TotalHardwareThread int32
	Capabilities        []string
	Cores               []*MachineCPUCoreBo
}

type MachineCPUBo struct {
	TotalCores          int32
	TotalHardwareThread int32
	Processors          []*MachineCPUProcessorBo
}

type MachineMemoryBo struct {
	TotalPhysicalBytes uint64
	TotalUsableBytes   uint64
	SupportedPageSizes []uint64
	UsedBytes          uint64
	FreeBytes          uint64
	SharedBytes        uint64
	BuffCacheBytes     uint64
	AvailableBytes     uint64
	SwapTotalBytes     uint64
	SwapUsedBytes      uint64
	SwapFreeBytes      uint64
}

type MachineDiskMountBo struct {
	MountPoint string
	FSType     string
	TotalBytes uint64
	UsedBytes  uint64
	FreeBytes  uint64
	FreeRate   float64
}

type MachineDiskBo struct {
	Name         string
	Type         string
	SizeBytes    uint64
	Vendor       string
	Model        string
	SerialNumber string
	WWN          string
	Mounts       []*MachineDiskMountBo
}

type MachineNetworkBo struct {
	LocalIP      string
	OutboundIP   string
	CIDR         string
	DNSServers   []string
	TotalRXBytes uint64
	TotalTXBytes uint64
	NICs         []string
}

type MachineInfoBo struct {
	HostName    string
	MachineUUID string
	CPU         *MachineCPUBo
	Memory      *MachineMemoryBo
	Disks       []*MachineDiskBo
	Network     *MachineNetworkBo
	System      *MachineSystemBo
}

type MachineSystemBo struct {
	Arch    string
	OS      string
	Version string
	Kernel  string
}

func ToAPIV1MachineInfoReply(in *MachineInfoBo) *apiv1.GetMachineInfoReply {
	if in == nil {
		return &apiv1.GetMachineInfoReply{}
	}
	out := &apiv1.GetMachineInfoReply{
		Host: &apiv1.HostInfo{
			HostName:    in.HostName,
			MachineUuid: in.MachineUUID,
		},
	}
	if in.CPU != nil {
		cpu := &apiv1.CPUInfo{
			TotalCores:           in.CPU.TotalCores,
			TotalHardwareThreads: in.CPU.TotalHardwareThread,
		}
		for _, p := range in.CPU.Processors {
			item := &apiv1.CPUProcessorItem{
				Id:                   p.ID,
				Vendor:               p.Vendor,
				Model:                p.Model,
				TotalCores:           p.TotalCores,
				TotalHardwareThreads: p.TotalHardwareThread,
				Capabilities:         p.Capabilities,
			}
			for _, c := range p.Cores {
				item.Cores = append(item.Cores, &apiv1.CPUCoreItem{
					Id:                c.ID,
					HardwareThreads:   c.HardwareThreads,
					LogicalProcessors: c.LogicalCPU,
				})
			}
			cpu.Processors = append(cpu.Processors, item)
		}
		out.Cpu = cpu
	}
	if in.Memory != nil {
		out.Memory = &apiv1.MemoryInfo{
			TotalPhysicalBytes: in.Memory.TotalPhysicalBytes,
			TotalUsableBytes:   in.Memory.TotalUsableBytes,
			SupportedPageSizes: in.Memory.SupportedPageSizes,
			UsedBytes:          in.Memory.UsedBytes,
			FreeBytes:          in.Memory.FreeBytes,
			SharedBytes:        in.Memory.SharedBytes,
			BuffCacheBytes:     in.Memory.BuffCacheBytes,
			AvailableBytes:     in.Memory.AvailableBytes,
			SwapTotalBytes:     in.Memory.SwapTotalBytes,
			SwapUsedBytes:      in.Memory.SwapUsedBytes,
			SwapFreeBytes:      in.Memory.SwapFreeBytes,
		}
	}
	for _, d := range in.Disks {
		item := &apiv1.DiskInfoItem{
			Name:         d.Name,
			Type:         d.Type,
			SizeBytes:    d.SizeBytes,
			Vendor:       d.Vendor,
			Model:        d.Model,
			SerialNumber: d.SerialNumber,
			Wwn:          d.WWN,
		}
		for _, m := range d.Mounts {
			item.Mounts = append(item.Mounts, &apiv1.DiskMountItem{
				MountPoint: m.MountPoint,
				FsType:     m.FSType,
				TotalBytes: m.TotalBytes,
				UsedBytes:  m.UsedBytes,
				FreeBytes:  m.FreeBytes,
				FreeRate:   m.FreeRate,
			})
		}
		out.Disks = append(out.Disks, item)
	}
	if in.Network != nil {
		out.Network = &apiv1.NetworkInfo{
			LocalIp:      in.Network.LocalIP,
			OutboundIp:   in.Network.OutboundIP,
			Cidr:         in.Network.CIDR,
			DnsServers:   in.Network.DNSServers,
			TotalRxBytes: in.Network.TotalRXBytes,
			TotalTxBytes: in.Network.TotalTXBytes,
			Nics:         in.Network.NICs,
		}
	}
	if in.System != nil {
		out.System = &apiv1.SystemInfo{
			Arch:    in.System.Arch,
			Os:      in.System.OS,
			Version: in.System.Version,
			Kernel:  in.System.Kernel,
		}
	}
	return out
}
