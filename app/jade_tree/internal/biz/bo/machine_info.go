package bo

import (
	"strings"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
	"github.com/aide-family/jade_tree/pkg/machine"
)

// MachineInfoIdentityBo is the natural key for a stored machine_infos row (machine UUID + hostname + local IP).
type MachineInfoIdentityBo struct {
	MachineUUID string
	HostName    string
	LocalIP     string
}

// DedupKey returns a stable string suitable for deduplicating payloads in memory.
func (b *MachineInfoIdentityBo) DedupKey() string {
	if b == nil {
		return ""
	}
	return b.MachineUUID + "\x1e" + b.HostName + "\x1e" + b.LocalIP
}

// NewMachineInfoIdentityBo builds a storage identity from collected or reported machine info.
func NewMachineInfoIdentityBo(in *machine.MachineInfo) *MachineInfoIdentityBo {
	if in == nil {
		return nil
	}
	lip := ""
	if in.Network != nil {
		lip = in.Network.LocalIP
	}
	return &MachineInfoIdentityBo{
		MachineUUID: in.MachineUUID,
		HostName:    in.HostName,
		LocalIP:     lip,
	}
}

type ListMachineInfosBo struct {
	*PageRequestBo
	Keywords string
	IP       string
	Hostname string
}

func NewListMachineInfosBo(req *apiv1.GetClusterMachineInfosRequest) *ListMachineInfosBo {
	page := int32(0)
	pageSize := int32(0)
	keywords := ""
	ip := ""
	hostname := ""
	if req != nil {
		page = req.GetPage()
		pageSize = req.GetPageSize()
		keywords = req.GetKeywords()
		ip = req.GetIp()
		hostname = req.GetHostname()
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	return &ListMachineInfosBo{
		PageRequestBo: NewPageRequestBo(page, pageSize),
		Keywords:      strings.TrimSpace(keywords),
		IP:            strings.TrimSpace(ip),
		Hostname:      strings.TrimSpace(hostname),
	}
}

func ToAPIV1MachineInfoReply(in *machine.MachineInfo) *apiv1.GetMachineInfoReply {
	if in == nil {
		return &apiv1.GetMachineInfoReply{}
	}
	out := &apiv1.GetMachineInfoReply{
		Host: &apiv1.HostInfo{
			HostName:    in.HostName,
			MachineUuid: in.MachineUUID,
		},
	}
	if cpuInfo := in.CPU; cpuInfo != nil {
		cpu := &apiv1.CPUInfo{
			TotalCores:           cpuInfo.TotalCores,
			TotalHardwareThreads: cpuInfo.TotalHardwareThread,
		}
		for _, p := range cpuInfo.Processors {
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
	if memoryInfo := in.Memory; memoryInfo != nil {
		out.Memory = &apiv1.MemoryInfo{
			TotalPhysicalBytes: memoryInfo.TotalPhysicalBytes,
			TotalUsableBytes:   memoryInfo.TotalUsableBytes,
			SupportedPageSizes: memoryInfo.SupportedPageSizes,
			UsedBytes:          memoryInfo.UsedBytes,
			FreeBytes:          memoryInfo.FreeBytes,
			SharedBytes:        memoryInfo.SharedBytes,
			BuffCacheBytes:     memoryInfo.BuffCacheBytes,
			AvailableBytes:     memoryInfo.AvailableBytes,
			SwapTotalBytes:     memoryInfo.SwapTotalBytes,
			SwapUsedBytes:      memoryInfo.SwapUsedBytes,
			SwapFreeBytes:      memoryInfo.SwapFreeBytes,
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
	if networkInfo := in.Network; networkInfo != nil {
		out.Network = &apiv1.NetworkInfo{
			LocalIp:      networkInfo.LocalIP,
			OutboundIp:   networkInfo.OutboundIP,
			Cidr:         networkInfo.CIDR,
			DnsServers:   networkInfo.DNSServers,
			TotalRxBytes: networkInfo.TotalRXBytes,
			TotalTxBytes: networkInfo.TotalTXBytes,
			Nics:         networkInfo.NICs,
		}
	}
	if systemInfo := in.System; systemInfo != nil {
		out.System = &apiv1.SystemInfo{
			Arch:    systemInfo.Arch,
			Os:      systemInfo.OS,
			Version: systemInfo.Version,
			Kernel:  systemInfo.Kernel,
		}
	}
	return out
}

func FromAPIV1MachineInfoReply(in *apiv1.GetMachineInfoReply) *machine.MachineInfo {
	if in == nil {
		return nil
	}

	out := &machine.MachineInfo{}
	if hostInfo := in.GetHost(); hostInfo != nil {
		out.HostName = hostInfo.GetHostName()
		out.MachineUUID = hostInfo.GetMachineUuid()
	}

	if cpuInfo := in.GetCpu(); cpuInfo != nil {
		cpu := &machine.MachineCPU{
			TotalCores:          cpuInfo.GetTotalCores(),
			TotalHardwareThread: cpuInfo.GetTotalHardwareThreads(),
		}
		for _, p := range cpuInfo.GetProcessors() {
			if p == nil {
				continue
			}
			pbo := &machine.MachineCPUProcessor{
				ID:                  p.GetId(),
				Vendor:              p.GetVendor(),
				Model:               p.GetModel(),
				TotalCores:          p.GetTotalCores(),
				TotalHardwareThread: p.GetTotalHardwareThreads(),
				Capabilities:        p.GetCapabilities(),
			}
			for _, c := range p.GetCores() {
				if c == nil {
					continue
				}
				pbo.Cores = append(pbo.Cores, &machine.MachineCPUCore{
					ID:              c.GetId(),
					HardwareThreads: c.GetHardwareThreads(),
					LogicalCPU:      c.GetLogicalProcessors(),
				})
			}
			cpu.Processors = append(cpu.Processors, pbo)
		}
		out.CPU = cpu
	}

	if memoryInfo := in.GetMemory(); memoryInfo != nil {
		out.Memory = &machine.MachineMemory{
			TotalPhysicalBytes: memoryInfo.GetTotalPhysicalBytes(),
			TotalUsableBytes:   memoryInfo.GetTotalUsableBytes(),
			SupportedPageSizes: memoryInfo.GetSupportedPageSizes(),
			UsedBytes:          memoryInfo.GetUsedBytes(),
			FreeBytes:          memoryInfo.GetFreeBytes(),
			SharedBytes:        memoryInfo.GetSharedBytes(),
			BuffCacheBytes:     memoryInfo.GetBuffCacheBytes(),
			AvailableBytes:     memoryInfo.GetAvailableBytes(),
			SwapTotalBytes:     memoryInfo.GetSwapTotalBytes(),
			SwapUsedBytes:      memoryInfo.GetSwapUsedBytes(),
			SwapFreeBytes:      memoryInfo.GetSwapFreeBytes(),
		}
	}

	for _, d := range in.GetDisks() {
		if d == nil {
			continue
		}
		item := &machine.MachineDisk{
			Name:         d.GetName(),
			Type:         d.GetType(),
			SizeBytes:    d.GetSizeBytes(),
			Vendor:       d.GetVendor(),
			Model:        d.GetModel(),
			SerialNumber: d.GetSerialNumber(),
			WWN:          d.GetWwn(),
		}
		for _, m := range d.GetMounts() {
			if m == nil {
				continue
			}
			item.Mounts = append(item.Mounts, &machine.MachineDiskMount{
				MountPoint: m.GetMountPoint(),
				FSType:     m.GetFsType(),
				TotalBytes: m.GetTotalBytes(),
				UsedBytes:  m.GetUsedBytes(),
				FreeBytes:  m.GetFreeBytes(),
				FreeRate:   m.GetFreeRate(),
			})
		}
		out.Disks = append(out.Disks, item)
	}

	if networkInfo := in.GetNetwork(); networkInfo != nil {
		out.Network = &machine.MachineNetwork{
			LocalIP:      networkInfo.GetLocalIp(),
			OutboundIP:   networkInfo.GetOutboundIp(),
			CIDR:         networkInfo.GetCidr(),
			DNSServers:   networkInfo.GetDnsServers(),
			TotalRXBytes: networkInfo.GetTotalRxBytes(),
			TotalTXBytes: networkInfo.GetTotalTxBytes(),
			NICs:         networkInfo.GetNics(),
		}
	}

	if systemInfo := in.GetSystem(); systemInfo != nil {
		out.System = &machine.MachineSystem{
			Arch:    systemInfo.GetArch(),
			OS:      systemInfo.GetOs(),
			Version: systemInfo.GetVersion(),
			Kernel:  systemInfo.GetKernel(),
		}
	}

	return out
}
