// Package impl provides repository implementations.
package impl

import (
	"bufio"
	"context"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"

	"github.com/denisbrodbeck/machineid"
	"github.com/jaypipes/ghw"
	ghost "github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"
	gnet "github.com/shirou/gopsutil/v3/net"

	"github.com/aide-family/jade_tree/internal/biz/bo"
	"github.com/aide-family/jade_tree/internal/biz/repository"
	"github.com/aide-family/jade_tree/internal/data"
)

func NewMachineInfoRepository(d *data.Data) repository.MachineInfoProvider {
	return &machineInfoRepository{Data: d}
}

type machineInfoRepository struct {
	*data.Data
}

func (m *machineInfoRepository) Collect(ctx context.Context) (*bo.MachineInfoBo, error) {
	hostName, _ := os.Hostname()
	out := &bo.MachineInfoBo{
		HostName:    hostName,
		MachineUUID: machineUUID(),
	}

	cpu, _ := ghw.CPU()
	if cpu != nil {
		cpuBo := &bo.MachineCPUBo{
			TotalCores:          int32(cpu.TotalCores),
			TotalHardwareThread: int32(cpu.TotalHardwareThreads),
		}
		for _, p := range cpu.Processors {
			if p == nil {
				continue
			}
			pbo := &bo.MachineCPUProcessorBo{
				ID:                  int32(p.ID),
				Vendor:              p.Vendor,
				Model:               p.Model,
				TotalCores:          int32(p.TotalCores),
				TotalHardwareThread: int32(p.TotalHardwareThreads),
				Capabilities:        p.Capabilities,
			}
			for _, c := range p.Cores {
				if c == nil {
					continue
				}
				logical := make([]int32, 0, len(c.LogicalProcessors))
				for _, lp := range c.LogicalProcessors {
					logical = append(logical, int32(lp))
				}
				pbo.Cores = append(pbo.Cores, &bo.MachineCPUCoreBo{
					ID:              int32(c.ID),
					HardwareThreads: int32(c.TotalHardwareThreads),
					LogicalCPU:      logical,
				})
			}
			cpuBo.Processors = append(cpuBo.Processors, pbo)
		}
		out.CPU = cpuBo
	}

	mem, _ := ghw.Memory()
	out.Memory = toMemoryInfo(mem)

	mountByDisk := getMountsByDisk()
	block, _ := ghw.Block()
	if block != nil {
		for _, d := range block.Disks {
			item := &bo.MachineDiskBo{
				Name:         d.Name,
				Type:         d.DriveType.String(),
				SizeBytes:    uint64(d.SizeBytes),
				Vendor:       d.Vendor,
				Model:        d.Model,
				SerialNumber: d.SerialNumber,
				WWN:          d.WWN,
			}
			if mounts, ok := mountByDisk[d.Name]; ok {
				item.Mounts = append(item.Mounts, mounts...)
			}
			out.Disks = append(out.Disks, item)
		}
	}

	out.Network = collectNetworkInfo()
	out.System = collectSystemInfo()
	_ = ctx
	return out, nil
}

func machineUUID() string {
	id, _ := machineid.ID()
	return id
}

func collectSystemInfo() *bo.MachineSystemBo {
	info, err := ghost.Info()
	if err == nil && info != nil {
		return &bo.MachineSystemBo{
			Arch:    info.KernelArch,
			OS:      info.OS,
			Version: info.PlatformVersion,
			Kernel:  info.KernelVersion,
		}
	}
	return &bo.MachineSystemBo{
		Arch: runtime.GOARCH,
		OS:   runtime.GOOS,
	}
}

func toMemoryInfo(info *ghw.MemoryInfo) *bo.MachineMemoryBo {
	vm, _ := mem.VirtualMemory()
	swap, _ := mem.SwapMemory()

	if info != nil {
		pageSizes := make([]uint64, 0, len(info.SupportedPageSizes))
		for _, p := range info.SupportedPageSizes {
			pageSizes = append(pageSizes, uint64(p))
		}
		return &bo.MachineMemoryBo{
			TotalPhysicalBytes: uint64(info.TotalPhysicalBytes),
			TotalUsableBytes:   uint64(info.TotalUsableBytes),
			SupportedPageSizes: pageSizes,
			UsedBytes:          getVMUint(vm, "used"),
			FreeBytes:          getVMUint(vm, "free"),
			SharedBytes:        getVMUint(vm, "shared"),
			BuffCacheBytes:     getVMUint(vm, "buffCache"),
			AvailableBytes:     getVMUint(vm, "available"),
			SwapTotalBytes:     getSwapUint(swap, "total"),
			SwapUsedBytes:      getSwapUint(swap, "used"),
			SwapFreeBytes:      getSwapUint(swap, "free"),
		}
	}
	if runtime.GOOS == "darwin" {
		size, err := readDarwinMemSize()
		if err == nil && size > 0 {
			return &bo.MachineMemoryBo{
				TotalPhysicalBytes: size,
				TotalUsableBytes:   size,
				UsedBytes:          getVMUint(vm, "used"),
				FreeBytes:          getVMUint(vm, "free"),
				SharedBytes:        getVMUint(vm, "shared"),
				BuffCacheBytes:     getVMUint(vm, "buffCache"),
				AvailableBytes:     getVMUint(vm, "available"),
				SwapTotalBytes:     getSwapUint(swap, "total"),
				SwapUsedBytes:      getSwapUint(swap, "used"),
				SwapFreeBytes:      getSwapUint(swap, "free"),
			}
		}
	}
	return nil
}

func getVMUint(vm *mem.VirtualMemoryStat, key string) uint64 {
	if vm == nil {
		return 0
	}
	switch key {
	case "used":
		return vm.Used
	case "free":
		return vm.Free
	case "shared":
		return vm.Shared
	case "available":
		return vm.Available
	case "buffCache":
		return vm.Buffers + vm.Cached
	default:
		return 0
	}
}

func getSwapUint(s *mem.SwapMemoryStat, key string) uint64 {
	if s == nil {
		return 0
	}
	switch key {
	case "total":
		return s.Total
	case "used":
		return s.Used
	case "free":
		return s.Free
	default:
		return 0
	}
}

func readDarwinMemSize() (uint64, error) {
	out, err := exec.Command("sysctl", "-n", "hw.memsize").Output()
	if err != nil {
		return 0, err
	}
	size, err := strconv.ParseUint(strings.TrimSpace(string(out)), 10, 64)
	if err != nil {
		return 0, err
	}
	return size, nil
}

func getMountsByDisk() map[string][]*bo.MachineDiskMountBo {
	result := make(map[string][]*bo.MachineDiskMountBo)
	out, err := exec.Command("df", "-kP").Output()
	if err != nil {
		return result
	}
	sc := bufio.NewScanner(strings.NewReader(string(out)))
	first := true
	for sc.Scan() {
		if first {
			first = false
			continue
		}
		parts := strings.Fields(sc.Text())
		if len(parts) < 6 {
			continue
		}
		device := parts[0]
		if !strings.HasPrefix(device, "/dev/") {
			continue
		}
		totalKB, err1 := strconv.ParseUint(parts[1], 10, 64)
		usedKB, err2 := strconv.ParseUint(parts[2], 10, 64)
		freeKB, err3 := strconv.ParseUint(parts[3], 10, 64)
		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}
		mountPoint := parts[5]
		total := totalKB * 1024
		used := usedKB * 1024
		free := freeKB * 1024
		rate := 0.0
		if total > 0 {
			rate = float64(free) / float64(total)
		}
		diskName := normalizeDiskName(device)
		if diskName == "" {
			continue
		}
		result[diskName] = append(result[diskName], &bo.MachineDiskMountBo{
			MountPoint: mountPoint,
			TotalBytes: total,
			UsedBytes:  used,
			FreeBytes:  free,
			FreeRate:   rate,
		})
	}
	return result
}

var darwinDiskRe = regexp.MustCompile(`^(disk\d+)`)

func normalizeDiskName(device string) string {
	base := filepath.Base(device)
	if strings.HasPrefix(base, "disk") {
		m := darwinDiskRe.FindStringSubmatch(base)
		if len(m) == 2 {
			return m[1]
		}
		return base
	}
	for len(base) > 0 {
		last := base[len(base)-1]
		if last < '0' || last > '9' {
			break
		}
		base = base[:len(base)-1]
	}
	return base
}

func collectNetworkInfo() *bo.MachineNetworkBo {
	n := &bo.MachineNetworkBo{}
	localIP, cidr := localIPAndCIDR()
	n.LocalIP = localIP
	n.CIDR = cidr
	n.OutboundIP = outboundIP()
	n.DNSServers = dnsServers()
	n.NICs, n.TotalRXBytes, n.TotalTXBytes = nicStats()
	return n
}

func localIPAndCIDR() (string, string) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP == nil || ipNet.IP.To4() == nil {
				continue
			}
			return ipNet.IP.String(), ipNet.String()
		}
	}
	return "", ""
}

func outboundIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer conn.Close()
	if addr, ok := conn.LocalAddr().(*net.UDPAddr); ok && addr.IP != nil {
		return addr.IP.String()
	}
	return ""
}

func dnsServers() []string {
	f, err := os.Open("/etc/resolv.conf")
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "nameserver ") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				out = append(out, parts[1])
			}
		}
	}
	return out
}

func nicStats() ([]string, uint64, uint64) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, 0, 0
	}
	nics := make([]string, 0)
	var rxTotal, txTotal uint64

	// Cross-platform NIC traffic counters.
	byName := make(map[string]gnet.IOCountersStat)
	if counters, err := gnet.IOCounters(true); err == nil {
		for _, c := range counters {
			byName[c.Name] = c
		}
	}

	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		nics = append(nics, iface.Name)
		if c, ok := byName[iface.Name]; ok {
			rxTotal += c.BytesRecv
			txTotal += c.BytesSent
			continue
		}
		// Linux fallback path when gopsutil counter is unavailable.
		rxTotal += readUint(filepath.Join("/sys/class/net", iface.Name, "statistics/rx_bytes"))
		txTotal += readUint(filepath.Join("/sys/class/net", iface.Name, "statistics/tx_bytes"))
	}
	return nics, rxTotal, txTotal
}

func readUint(path string) uint64 {
	b, err := os.ReadFile(path)
	if err != nil {
		return 0
	}
	v, _ := strconv.ParseUint(strings.TrimSpace(string(b)), 10, 64)
	return v
}
