package machine

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

// renderMachineDetail renders one section (cpu|memory|network|disk) for a single endpoint response.
func renderMachineDetail(kind, format, endpoint string, reply *apiv1.GetMachineInfoReply) error {
	if reply == nil {
		reply = &apiv1.GetMachineInfoReply{}
	}
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "cpu":
		return renderCPUDetail(format, endpoint, reply)
	case "memory":
		return renderMemoryDetail(format, endpoint, reply)
	case "network":
		return renderNetworkDetail(format, endpoint, reply)
	case "disk":
		return renderDiskDetail(format, endpoint, reply)
	default:
		return fmt.Errorf("unknown detail kind: %s", kind)
	}
}

func renderMachineDetailsMulti(kind, format string, items []machineDetailItem) error {
	if len(items) == 0 {
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		wrapped := make([]map[string]any, 0, len(items))
		for _, it := range items {
			wrapped = append(wrapped, detailPayloadAsMap(kind, it.Endpoint, it.Reply))
		}
		b, err := json.MarshalIndent(wrapped, "", "  ")
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(append(b, '\n'))
		return nil
	case "yaml":
		wrapped := make([]map[string]any, 0, len(items))
		for _, it := range items {
			wrapped = append(wrapped, detailPayloadAsMap(kind, it.Endpoint, it.Reply))
		}
		b, err := yaml.Marshal(wrapped)
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(b)
		return nil
	default:
		for i, it := range items {
			if i > 0 {
				_, _ = fmt.Fprintln(os.Stdout)
			}
			_, _ = fmt.Fprintf(os.Stdout, "--- endpoint: %s ---\n", it.Endpoint)
			if err := renderMachineDetail(kind, "table", it.Endpoint, it.Reply); err != nil {
				return err
			}
		}
		return nil
	}
}

type machineDetailItem struct {
	Endpoint string
	Reply    *apiv1.GetMachineInfoReply
}

func detailPayloadAsMap(kind, endpoint string, reply *apiv1.GetMachineInfoReply) map[string]any {
	if reply == nil {
		reply = &apiv1.GetMachineInfoReply{}
	}
	out := map[string]any{"endpoint": endpoint, "hostName": reply.GetHost().GetHostName()}
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "cpu":
		out["cpu"] = cpuToMap(reply.GetCpu())
	case "memory":
		out["memory"] = memoryToMap(reply.GetMemory())
	case "network":
		out["network"] = networkToMap(reply.GetNetwork())
	case "disk":
		out["disks"] = disksToSlice(reply.GetDisks())
	}
	return out
}

func cpuToMap(c *apiv1.CPUInfo) map[string]any {
	if c == nil {
		return map[string]any{}
	}
	procs := make([]map[string]any, 0, len(c.GetProcessors()))
	for _, p := range c.GetProcessors() {
		if p == nil {
			continue
		}
		cores := make([]map[string]any, 0, len(p.GetCores()))
		for _, co := range p.GetCores() {
			if co == nil {
				continue
			}
			cores = append(cores, map[string]any{
				"id":                co.GetId(),
				"hardwareThreads":   co.GetHardwareThreads(),
				"logicalProcessors": co.GetLogicalProcessors(),
			})
		}
		procs = append(procs, map[string]any{
			"id":                   p.GetId(),
			"vendor":               p.GetVendor(),
			"model":                p.GetModel(),
			"totalCores":           p.GetTotalCores(),
			"totalHardwareThreads": p.GetTotalHardwareThreads(),
			"capabilities":         p.GetCapabilities(),
			"cores":                cores,
		})
	}
	return map[string]any{
		"totalCores":           c.GetTotalCores(),
		"totalHardwareThreads": c.GetTotalHardwareThreads(),
		"processors":           procs,
	}
}

func memoryToMap(m *apiv1.MemoryInfo) map[string]any {
	if m == nil {
		return map[string]any{}
	}
	return map[string]any{
		"totalPhysicalBytes": m.GetTotalPhysicalBytes(),
		"totalUsableBytes":   m.GetTotalUsableBytes(),
		"supportedPageSizes": m.GetSupportedPageSizes(),
		"usedBytes":          m.GetUsedBytes(),
		"freeBytes":          m.GetFreeBytes(),
		"sharedBytes":        m.GetSharedBytes(),
		"buffCacheBytes":     m.GetBuffCacheBytes(),
		"availableBytes":     m.GetAvailableBytes(),
		"swapTotalBytes":     m.GetSwapTotalBytes(),
		"swapUsedBytes":      m.GetSwapUsedBytes(),
		"swapFreeBytes":      m.GetSwapFreeBytes(),
	}
}

func networkToMap(n *apiv1.NetworkInfo) map[string]any {
	if n == nil {
		return map[string]any{}
	}
	return map[string]any{
		"localIp":      n.GetLocalIp(),
		"outboundIp":   n.GetOutboundIp(),
		"cidr":         n.GetCidr(),
		"dnsServers":   n.GetDnsServers(),
		"totalRxBytes": n.GetTotalRxBytes(),
		"totalTxBytes": n.GetTotalTxBytes(),
		"nics":         n.GetNics(),
	}
}

func disksToSlice(disks []*apiv1.DiskInfoItem) []map[string]any {
	out := make([]map[string]any, 0, len(disks))
	for _, d := range disks {
		if d == nil {
			continue
		}
		mounts := make([]map[string]any, 0, len(d.GetMounts()))
		for _, m := range d.GetMounts() {
			if m == nil {
				continue
			}
			mounts = append(mounts, map[string]any{
				"mountPoint": m.GetMountPoint(),
				"fsType":     m.GetFsType(),
				"totalBytes": m.GetTotalBytes(),
				"usedBytes":  m.GetUsedBytes(),
				"freeBytes":  m.GetFreeBytes(),
				"freeRate":   m.GetFreeRate(),
			})
		}
		out = append(out, map[string]any{
			"name":         d.GetName(),
			"type":         d.GetType(),
			"sizeBytes":    d.GetSizeBytes(),
			"vendor":       d.GetVendor(),
			"model":        d.GetModel(),
			"serialNumber": d.GetSerialNumber(),
			"wwn":          d.GetWwn(),
			"mounts":       mounts,
		})
	}
	return out
}

func renderCPUDetail(format, endpoint string, reply *apiv1.GetMachineInfoReply) error {
	payload := detailPayloadAsMap("cpu", endpoint, reply)
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(payload)
	case "yaml":
		return writeYAML(payload)
	default:
		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetHeader([]string{"FIELD", "VALUE"})
		host := reply.GetHost().GetHostName()
		if host != "" {
			tw.Append([]string{"HOST", host})
		}
		if ep := strings.TrimSpace(endpoint); ep != "" {
			tw.Append([]string{"ENDPOINT", ep})
		}
		cpu := reply.GetCpu()
		if cpu == nil {
			tw.Append([]string{"CPU", "(none)"})
			tw.Render()
			return nil
		}
		tw.Append([]string{"TOTAL_CORES", strconv.FormatInt(int64(cpu.GetTotalCores()), 10)})
		tw.Append([]string{"TOTAL_HW_THREADS", strconv.FormatInt(int64(cpu.GetTotalHardwareThreads()), 10)})
		tw.Render()
		for _, p := range cpu.GetProcessors() {
			if p == nil {
				continue
			}
			_, _ = fmt.Fprintln(os.Stdout)
			_, _ = fmt.Fprintf(os.Stdout, "Processor %d: %s %s\n", p.GetId(), p.GetVendor(), p.GetModel())
			pt := tablewriter.NewWriter(os.Stdout)
			pt.SetHeader([]string{"FIELD", "VALUE"})
			pt.Append([]string{"TOTAL_CORES", strconv.FormatInt(int64(p.GetTotalCores()), 10)})
			pt.Append([]string{"TOTAL_HW_THREADS", strconv.FormatInt(int64(p.GetTotalHardwareThreads()), 10)})
			if len(p.GetCapabilities()) > 0 {
				pt.Append([]string{"CAPABILITIES", strings.Join(p.GetCapabilities(), ", ")})
			}
			pt.Render()
			if len(p.GetCores()) == 0 {
				continue
			}
			ct := tablewriter.NewWriter(os.Stdout)
			ct.SetHeader([]string{"CORE_ID", "HW_THREADS", "LOGICAL_CPUS"})
			for _, c := range p.GetCores() {
				if c == nil {
					continue
				}
				logical := intsToString(c.GetLogicalProcessors())
				ct.Append([]string{
					strconv.FormatInt(int64(c.GetId()), 10),
					strconv.FormatInt(int64(c.GetHardwareThreads()), 10),
					logical,
				})
			}
			ct.Render()
		}
		return nil
	}
}

func renderMemoryDetail(format, endpoint string, reply *apiv1.GetMachineInfoReply) error {
	payload := detailPayloadAsMap("memory", endpoint, reply)
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(payload)
	case "yaml":
		return writeYAML(payload)
	default:
		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetHeader([]string{"FIELD", "VALUE"})
		if host := reply.GetHost().GetHostName(); host != "" {
			tw.Append([]string{"HOST", host})
		}
		if ep := strings.TrimSpace(endpoint); ep != "" {
			tw.Append([]string{"ENDPOINT", ep})
		}
		m := reply.GetMemory()
		if m == nil {
			tw.Append([]string{"MEMORY", "(none)"})
			tw.Render()
			return nil
		}
		rows := []struct{ k, v string }{
			{"TOTAL_PHYSICAL", formatBytes(m.GetTotalPhysicalBytes())},
			{"TOTAL_USABLE", formatBytes(m.GetTotalUsableBytes())},
			{"USED", formatBytes(m.GetUsedBytes())},
			{"FREE", formatBytes(m.GetFreeBytes())},
			{"AVAILABLE", formatBytes(m.GetAvailableBytes())},
			{"SHARED", formatBytes(m.GetSharedBytes())},
			{"BUFF_CACHE", formatBytes(m.GetBuffCacheBytes())},
			{"SWAP_TOTAL", formatBytes(m.GetSwapTotalBytes())},
			{"SWAP_USED", formatBytes(m.GetSwapUsedBytes())},
			{"SWAP_FREE", formatBytes(m.GetSwapFreeBytes())},
		}
		for _, row := range rows {
			tw.Append([]string{row.k, row.v})
		}
		if len(m.GetSupportedPageSizes()) > 0 {
			sizes := make([]string, 0, len(m.GetSupportedPageSizes()))
			for _, sz := range m.GetSupportedPageSizes() {
				sizes = append(sizes, formatBytes(sz))
			}
			tw.Append([]string{"PAGE_SIZES", strings.Join(sizes, ", ")})
		}
		tw.Render()
		return nil
	}
}

func renderNetworkDetail(format, endpoint string, reply *apiv1.GetMachineInfoReply) error {
	payload := detailPayloadAsMap("network", endpoint, reply)
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(payload)
	case "yaml":
		return writeYAML(payload)
	default:
		tw := tablewriter.NewWriter(os.Stdout)
		tw.SetHeader([]string{"FIELD", "VALUE"})
		if host := reply.GetHost().GetHostName(); host != "" {
			tw.Append([]string{"HOST", host})
		}
		if ep := strings.TrimSpace(endpoint); ep != "" {
			tw.Append([]string{"ENDPOINT", ep})
		}
		n := reply.GetNetwork()
		if n == nil {
			tw.Append([]string{"NETWORK", "(none)"})
			tw.Render()
			return nil
		}
		tw.Append([]string{"LOCAL_IP", n.GetLocalIp()})
		tw.Append([]string{"OUTBOUND_IP", n.GetOutboundIp()})
		tw.Append([]string{"CIDR", n.GetCidr()})
		tw.Append([]string{"DNS_SERVERS", strings.Join(n.GetDnsServers(), ", ")})
		tw.Append([]string{"NICS", strings.Join(n.GetNics(), ", ")})
		tw.Append([]string{"TOTAL_RX", formatBytes(n.GetTotalRxBytes())})
		tw.Append([]string{"TOTAL_TX", formatBytes(n.GetTotalTxBytes())})
		tw.Render()
		return nil
	}
}

func renderDiskDetail(format, endpoint string, reply *apiv1.GetMachineInfoReply) error {
	payload := detailPayloadAsMap("disk", endpoint, reply)
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		return writeJSON(payload)
	case "yaml":
		return writeYAML(payload)
	default:
		if host := reply.GetHost().GetHostName(); host != "" {
			_, _ = fmt.Fprintf(os.Stdout, "HOST: %s\n", host)
		}
		if ep := strings.TrimSpace(endpoint); ep != "" {
			_, _ = fmt.Fprintf(os.Stdout, "ENDPOINT: %s\n", ep)
		}
		disks := reply.GetDisks()
		if len(disks) == 0 {
			_, _ = fmt.Fprintln(os.Stdout, "(no disks)")
			return nil
		}
		for i, d := range disks {
			if d == nil {
				continue
			}
			if i > 0 {
				_, _ = fmt.Fprintln(os.Stdout)
			}
			_, _ = fmt.Fprintf(os.Stdout, "Disk: %s (%s)\n", d.GetName(), d.GetType())
			dt := tablewriter.NewWriter(os.Stdout)
			dt.SetHeader([]string{"FIELD", "VALUE"})
			dt.Append([]string{"SIZE", formatBytes(d.GetSizeBytes())})
			dt.Append([]string{"VENDOR", d.GetVendor()})
			dt.Append([]string{"MODEL", d.GetModel()})
			dt.Append([]string{"SERIAL", d.GetSerialNumber()})
			dt.Append([]string{"WWN", d.GetWwn()})
			dt.Render()
			if len(d.GetMounts()) == 0 {
				continue
			}
			mt := tablewriter.NewWriter(os.Stdout)
			mt.SetHeader([]string{"MOUNT", "FS", "TOTAL", "USED", "FREE", "FREE%"})
			for _, m := range d.GetMounts() {
				if m == nil {
					continue
				}
				pct := ""
				if m.GetFreeRate() > 0 {
					pct = fmt.Sprintf("%.1f%%", m.GetFreeRate()*100)
				}
				mt.Append([]string{
					m.GetMountPoint(),
					m.GetFsType(),
					formatBytes(m.GetTotalBytes()),
					formatBytes(m.GetUsedBytes()),
					formatBytes(m.GetFreeBytes()),
					pct,
				})
			}
			mt.Render()
		}
		return nil
	}
}

func writeJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	_, _ = os.Stdout.Write(append(b, '\n'))
	return nil
}

func writeYAML(v any) error {
	b, err := yaml.Marshal(v)
	if err != nil {
		return err
	}
	_, _ = os.Stdout.Write(b)
	return nil
}

func intsToString(xs []int32) string {
	if len(xs) == 0 {
		return ""
	}
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = strconv.FormatInt(int64(x), 10)
	}
	return strings.Join(parts, ",")
}
