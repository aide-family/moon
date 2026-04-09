package machine

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"

	"github.com/aide-family/magicbox/enum"
	"github.com/bwmarrin/snowflake"
)

var (
	_ sql.Scanner   = (*MachineInfo)(nil)
	_ driver.Valuer = (*MachineInfo)(nil)
)

type MachineInfo struct {
	ID          snowflake.ID
	HostName    string
	MachineUUID string
	Source      enum.MachineInfoSource
	Agent       *MachineAgent
	CPU         *MachineCPU
	Memory      *MachineMemory
	Disks       []*MachineDisk
	Network     *MachineNetwork
	System      *MachineSystem
}

type MachineAgent struct {
	Endpoint     string
	HTTPEndpoint string
	GRPCEndpoint string
	Version      string
}

// Value implements [driver.Valuer].
func (m *MachineInfo) Value() (driver.Value, error) {
	return json.Marshal(m)
}

// Scan implements [sql.Scanner].
func (m *MachineInfo) Scan(src any) error {
	if src == nil {
		return nil
	}
	switch v := src.(type) {
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	default:
		return fmt.Errorf("unsupported type: %T", src)
	}
}

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

type MachineSystem struct {
	Arch    string
	OS      string
	Version string
	Kernel  string
}
