package machine

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

const (
	DetailKindCPU     = "cpu"
	DetailKindMemory  = "memory"
	DetailKindNetwork = "network"
	DetailKindDisk    = "disk"
	DetailKindSystem  = "system"
)

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
		switch strings.ToLower(strings.TrimSpace(kind)) {
		case DetailKindCPU:
			return renderCPUDetailsComparisonTable(items)
		case DetailKindMemory:
			return renderMemoryDetailsComparisonTable(items)
		case DetailKindNetwork:
			return renderNetworkDetailsComparisonTable(items)
		case DetailKindDisk:
			return renderDiskDetailsComparisonTable(items)
		case DetailKindSystem:
			return renderSystemDetailsComparisonTable(items)
		default:
			return fmt.Errorf("unknown detail kind: %s", kind)
		}
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
	case DetailKindCPU:
		out[DetailKindCPU] = cpuToMap(reply.GetCpu())
	case DetailKindMemory:
		out[DetailKindMemory] = memoryToMap(reply.GetMemory())
	case DetailKindNetwork:
		out[DetailKindNetwork] = networkToMap(reply.GetNetwork())
	case DetailKindDisk:
		out[DetailKindDisk] = disksToSlice(reply.GetDisks())
	case DetailKindSystem:
		out[DetailKindSystem] = systemToMap(reply.GetSystem())
	}
	return out
}

func systemToMap(s *apiv1.SystemInfo) map[string]any {
	if s == nil {
		return map[string]any{}
	}
	return map[string]any{
		"arch":    s.GetArch(),
		"os":      s.GetOs(),
		"version": s.GetVersion(),
		"kernel":  s.GetKernel(),
	}
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

func replyOrEmpty(reply *apiv1.GetMachineInfoReply) *apiv1.GetMachineInfoReply {
	if reply == nil {
		return &apiv1.GetMachineInfoReply{}
	}
	return reply
}

func formatCPUCoresSummary(cores []*apiv1.CPUCoreItem) string {
	if len(cores) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(cores))
	for _, c := range cores {
		if c == nil {
			continue
		}
		parts = append(parts, fmt.Sprintf("%d:%dt:%s",
			c.GetId(), c.GetHardwareThreads(), intsToString(c.GetLogicalProcessors())))
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "; ")
}

func formatDiskMountsSummary(mounts []*apiv1.DiskMountItem) string {
	if len(mounts) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(mounts))
	for _, m := range mounts {
		if m == nil {
			continue
		}
		frag := fmt.Sprintf("%s [%s] %s/%s/%s",
			m.GetMountPoint(),
			m.GetFsType(),
			formatBytes(m.GetTotalBytes()),
			formatBytes(m.GetUsedBytes()),
			formatBytes(m.GetFreeBytes()),
		)
		if m.GetFreeRate() > 0 {
			frag += fmt.Sprintf(" %.1f%% free", m.GetFreeRate()*100)
		}
		parts = append(parts, frag)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "\n")
}

func maxLineRuneWidth(s string) int {
	m := 0
	for _, line := range strings.Split(s, "\n") {
		if n := utf8.RuneCountInString(line); n > m {
			m = n
		}
	}
	return m
}

// scanDiskMetaMountsLineWidths returns the widest single line (in runes) seen in META and MOUNTS cells across all disks.
func scanDiskMetaMountsLineWidths(items []machineDetailItem) (metaMax, mountMax int) {
	for _, it := range items {
		reply := replyOrEmpty(it.Reply)
		for _, d := range reply.GetDisks() {
			if d == nil {
				continue
			}
			if w := maxLineRuneWidth(formatDiskMetaMultiline(d)); w > metaMax {
				metaMax = w
			}
			if w := maxLineRuneWidth(formatDiskMountsSummary(d.GetMounts())); w > mountMax {
				mountMax = w
			}
		}
	}
	return metaMax, mountMax
}

// formatDiskMetaMultiline stacks vendor, model, serial, and WWN on separate lines to keep table width reasonable.
func formatDiskMetaMultiline(d *apiv1.DiskInfoItem) string {
	if d == nil {
		return "-"
	}
	var b strings.Builder
	appendLine := func(s string) {
		s = strings.TrimSpace(s)
		if s == "" {
			return
		}
		if b.Len() > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(s)
	}
	appendLine(d.GetVendor())
	appendLine(d.GetModel())
	appendLine(d.GetSerialNumber())
	appendLine(d.GetWwn())
	if b.Len() == 0 {
		return "-"
	}
	return b.String()
}

func formatPageSizesSummary(sizes []uint64) string {
	if len(sizes) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(sizes))
	for _, sz := range sizes {
		parts = append(parts, formatBytes(sz))
	}
	return strings.Join(parts, ", ")
}

// cpuCapabilitiesMaxLineRunes limits CAPABILITIES column width in the CPU detail table (table output only).
const cpuCapabilitiesMaxLineRunes = 56

// wrapCommaSeparatedToRunes breaks a comma-separated list into multiple lines so each line is at most maxRunes runes wide (breaks at commas; overlong tokens are hard-split).
func wrapCommaSeparatedToRunes(s string, maxRunes int) string {
	s = strings.TrimSpace(s)
	if s == "" || maxRunes <= 0 {
		return s
	}
	parts := strings.Split(s, ",")
	tokens := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			tokens = append(tokens, p)
		}
	}
	if len(tokens) == 0 {
		return s
	}

	hardSplit := func(tok string) []string {
		if utf8.RuneCountInString(tok) <= maxRunes {
			return []string{tok}
		}
		var out []string
		for len(tok) > 0 {
			var chunk strings.Builder
			n := 0
			for _, r := range tok {
				if n >= maxRunes {
					break
				}
				chunk.WriteRune(r)
				n++
			}
			out = append(out, chunk.String())
			tok = tok[len(chunk.String()):]
		}
		return out
	}

	var lines []string
	var b strings.Builder
	lineRunes := 0
	flushLine := func() {
		if b.Len() > 0 {
			lines = append(lines, b.String())
			b.Reset()
			lineRunes = 0
		}
	}
	for _, tok := range tokens {
		pieces := hardSplit(tok)
		for pi, piece := range pieces {
			pw := utf8.RuneCountInString(piece)
			if pi == 0 {
				if b.Len() > 0 {
					if lineRunes+1+pw > maxRunes {
						flushLine()
					} else {
						b.WriteByte(',')
						lineRunes++
					}
				}
			} else {
				if lineRunes+pw > maxRunes {
					flushLine()
				}
			}
			b.WriteString(piece)
			lineRunes += pw
		}
	}
	flushLine()
	return strings.Join(lines, "\n")
}

func renderCPUDetailsComparisonTable(items []machineDetailItem) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetAutoWrapText(true)
	tw.SetReflowDuringAutoWrap(false)
	tw.SetHeader([]string{
		"ENDPOINT", "HOSTNAME", "TOTAL_CORES", "TOTAL_HW_THREADS",
		"PROC_ID", "VENDOR", "MODEL", "PROC_CORES", "PROC_HW_THREADS", "CAPABILITIES", "CORES",
	})
	const colCapabilities = 9
	tw.SetColMinWidth(colCapabilities, cpuCapabilitiesMaxLineRunes)
	for _, it := range items {
		reply := replyOrEmpty(it.Reply)
		ep := strings.TrimSpace(it.Endpoint)
		host := reply.GetHost().GetHostName()
		cpu := reply.GetCpu()
		if cpu == nil {
			tw.Append([]string{ep, host, "-", "-", "-", "-", "-", "-", "-", "-", "-"})
			continue
		}
		tc := strconv.FormatInt(int64(cpu.GetTotalCores()), 10)
		tt := strconv.FormatInt(int64(cpu.GetTotalHardwareThreads()), 10)
		procs := cpu.GetProcessors()
		if len(procs) == 0 {
			tw.Append([]string{ep, host, tc, tt, "-", "-", "-", "-", "-", "-", "-"})
			continue
		}
		procRows := 0
		for _, p := range procs {
			if p == nil {
				continue
			}
			procRows++
			caps := strings.Join(p.GetCapabilities(), ",")
			if caps == "" {
				caps = "-"
			} else {
				caps = wrapCommaSeparatedToRunes(caps, cpuCapabilitiesMaxLineRunes)
			}
			tw.Append([]string{
				ep, host, tc, tt,
				strconv.FormatInt(int64(p.GetId()), 10),
				p.GetVendor(),
				p.GetModel(),
				strconv.FormatInt(int64(p.GetTotalCores()), 10),
				strconv.FormatInt(int64(p.GetTotalHardwareThreads()), 10),
				caps,
				formatCPUCoresSummary(p.GetCores()),
			})
		}
		if procRows == 0 {
			tw.Append([]string{ep, host, tc, tt, "-", "-", "-", "-", "-", "-", "-"})
		}
	}
	tw.Render()
	return nil
}

func renderMemoryDetailsComparisonTable(items []machineDetailItem) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetAutoWrapText(false)
	tw.SetHeader([]string{
		"ENDPOINT", "HOSTNAME", "TOTAL_PHYSICAL", "TOTAL_USABLE", "USED", "FREE", "AVAILABLE",
		"SHARED", "BUFF_CACHE", "SWAP_TOTAL", "SWAP_USED", "SWAP_FREE", "PAGE_SIZES",
	})
	for _, it := range items {
		reply := replyOrEmpty(it.Reply)
		ep := strings.TrimSpace(it.Endpoint)
		host := reply.GetHost().GetHostName()
		m := reply.GetMemory()
		if m == nil {
			tw.Append([]string{ep, host, "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-", "-"})
			continue
		}
		tw.Append([]string{
			ep, host,
			formatBytes(m.GetTotalPhysicalBytes()),
			formatBytes(m.GetTotalUsableBytes()),
			formatBytes(m.GetUsedBytes()),
			formatBytes(m.GetFreeBytes()),
			formatBytes(m.GetAvailableBytes()),
			formatBytes(m.GetSharedBytes()),
			formatBytes(m.GetBuffCacheBytes()),
			formatBytes(m.GetSwapTotalBytes()),
			formatBytes(m.GetSwapUsedBytes()),
			formatBytes(m.GetSwapFreeBytes()),
			formatPageSizesSummary(m.GetSupportedPageSizes()),
		})
	}
	tw.Render()
	return nil
}

// formatNetworkMetaMultiline stacks DNS servers and NIC names on separate lines (with short section labels)
// so the comparison table stays readable in narrow terminals.
func formatNetworkMetaMultiline(n *apiv1.NetworkInfo) string {
	if n == nil {
		return "-"
	}
	var parts []string
	var dnsLines []string
	for _, s := range n.GetDnsServers() {
		s = strings.TrimSpace(s)
		if s != "" {
			dnsLines = append(dnsLines, s)
		}
	}
	var nicLines []string
	for _, s := range n.GetNics() {
		s = strings.TrimSpace(s)
		if s != "" {
			nicLines = append(nicLines, s)
		}
	}
	if len(dnsLines) > 0 {
		parts = append(parts, "DNS:\n")
		parts = append(parts, dnsLines...)
	}
	if len(nicLines) > 0 {
		if len(parts) > 0 {
			parts = append(parts, "")
		}
		parts = append(parts, "NICs:\n")
		parts = append(parts, nicLines...)
	}
	if len(parts) == 0 {
		return "-"
	}
	return strings.Join(parts, "\n")
}

func renderNetworkDetailsComparisonTable(items []machineDetailItem) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetAutoWrapText(true)
	tw.SetHeader([]string{
		"ENDPOINT", "HOSTNAME", "LOCAL_IP", "OUTBOUND_IP", "CIDR", "TOTAL_RX", "TOTAL_TX", "META",
	})
	for _, it := range items {
		reply := replyOrEmpty(it.Reply)
		ep := strings.TrimSpace(it.Endpoint)
		host := reply.GetHost().GetHostName()
		n := reply.GetNetwork()
		if n == nil {
			tw.Append([]string{ep, host, "-", "-", "-", "-", "-", "-"})
			continue
		}
		tw.Append([]string{
			ep, host,
			n.GetLocalIp(),
			n.GetOutboundIp(),
			n.GetCidr(),
			formatBytes(n.GetTotalRxBytes()),
			formatBytes(n.GetTotalTxBytes()),
			formatNetworkMetaMultiline(n),
		})
	}
	tw.Render()
	return nil
}

func renderDiskDetailsComparisonTable(items []machineDetailItem) error {
	metaMax, mountMax := scanDiskMetaMountsLineWidths(items)
	const wrapFloor = 72
	wrapW := max(metaMax, mountMax, wrapFloor)
	metaColW := max(metaMax, utf8.RuneCountInString("META"))
	mountColW := max(mountMax, utf8.RuneCountInString("MOUNTS"))

	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetAutoWrapText(true)
	tw.SetReflowDuringAutoWrap(false)
	tw.SetColWidth(wrapW)
	tw.SetHeader([]string{
		"ENDPOINT", "HOSTNAME", "DISK", "TYPE", "SIZE", "META", "MOUNTS",
	})
	tw.SetColMinWidth(5, metaColW)
	tw.SetColMinWidth(6, mountColW)
	for _, it := range items {
		reply := replyOrEmpty(it.Reply)
		ep := strings.TrimSpace(it.Endpoint)
		host := reply.GetHost().GetHostName()
		disks := reply.GetDisks()
		if len(disks) == 0 {
			tw.Append([]string{ep, host, "-", "-", "-", "-", "-"})
			continue
		}
		diskRows := 0
		for _, d := range disks {
			if d == nil {
				continue
			}
			diskRows++
			tw.Append([]string{
				ep, host,
				d.GetName(),
				d.GetType(),
				formatBytes(d.GetSizeBytes()),
				formatDiskMetaMultiline(d),
				formatDiskMountsSummary(d.GetMounts()),
			})
		}
		if diskRows == 0 {
			tw.Append([]string{ep, host, "-", "-", "-", "-", "-"})
		}
	}
	tw.Render()
	return nil
}

func renderSystemDetailsComparisonTable(items []machineDetailItem) error {
	tw := tablewriter.NewWriter(os.Stdout)
	tw.SetAutoWrapText(true)
	tw.SetHeader([]string{
		"ENDPOINT", "HOSTNAME", "ARCH", "OS", "VERSION", "KERNEL",
	})
	for _, it := range items {
		reply := replyOrEmpty(it.Reply)
		ep := strings.TrimSpace(it.Endpoint)
		host := reply.GetHost().GetHostName()
		s := reply.GetSystem()
		if s == nil {
			tw.Append([]string{ep, host, "-", "-", "-", "-"})
			continue
		}
		dashIfEmpty := func(v string) string {
			v = strings.TrimSpace(v)
			if v == "" {
				return "-"
			}
			return v
		}
		tw.Append([]string{
			ep, host,
			dashIfEmpty(s.GetArch()),
			dashIfEmpty(s.GetOs()),
			dashIfEmpty(s.GetVersion()),
			dashIfEmpty(s.GetKernel()),
		})
	}
	tw.Render()
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
