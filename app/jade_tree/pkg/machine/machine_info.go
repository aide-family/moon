package machine

import (
	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"
)

type MachineCPUCore struct {
	ID              int32
	HardwareThreads int32
	LogicalCPU      []int32
}

type MachineCPUProcessor struct {
	ID                  int32
	Vendor              string
	Model               string
	TotalCores          int32
	TotalHardwareThread int32
	Capabilities        []string
	Cores               []*MachineCPUCore
}

type MachineCPU struct {
	TotalCores          int32
	TotalHardwareThread int32
	Processors          []*MachineCPUProcessor
}

type MachineMemory struct {
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

type MachineDiskMount struct {
	MountPoint string
	FSType     string
	TotalBytes uint64
	UsedBytes  uint64
	FreeBytes  uint64
	FreeRate   float64
}

type MachineDisk struct {
	Name         string
	Type         string
	SizeBytes    uint64
	Vendor       string
	Model        string
	SerialNumber string
	WWN          string
	Mounts       []*MachineDiskMount
}

type MachineNetwork struct {
	LocalIP      string
	OutboundIP   string
	CIDR         string
	DNSServers   []string
	TotalRXBytes uint64
	TotalTXBytes uint64
	NICs         []string
}

type MachineInfo struct {
	ID          snowflake.ID
	HostName    string
	MachineUUID string
	Source      enum.MachineInfoSource
	CPU         *MachineCPU
	Memory      *MachineMemory
	Disks       []*MachineDisk
	Network     *MachineNetwork
	System      *MachineSystem
}

type MachineSystem struct {
	Arch    string
	OS      string
	Version string
	Kernel  string
}
