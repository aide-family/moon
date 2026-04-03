package machine

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"gopkg.in/yaml.v3"

	apiv1 "github.com/aide-family/jade_tree/pkg/api/v1"
)

type machineRow struct {
	Endpoint      string `json:"endpoint" yaml:"endpoint"`
	HostName      string `json:"hostName" yaml:"hostName"`
	LocalIP       string `json:"ip" yaml:"ip"`
	CPUName       string `json:"cpuName" yaml:"cpuName"`
	CPUCores      int32  `json:"cpuCores" yaml:"cpuCores"`
	MemoryTotal   string `json:"memoryTotal" yaml:"memoryTotal"`
	MemoryAvail   string `json:"memoryAvailable" yaml:"memoryAvailable"`
	DiskTotal     string `json:"diskTotal" yaml:"diskTotal"`
	DiskAvailable string `json:"diskAvailable" yaml:"diskAvailable"`
	NetworkRx     string `json:"networkRx" yaml:"networkRx"`
	NetworkTx     string `json:"networkTx" yaml:"networkTx"`

	// Display-only fields used by table output.
	// These are unexported to keep JSON/YAML shape stable.
	endpointCell string
	hostNameCell string
}

func toMachineRows(endpoint string, machines []*apiv1.GetMachineInfoReply) []machineRow {
	rows := make([]machineRow, 0, len(machines))
	endpoint = strings.TrimSpace(endpoint)
	for _, item := range machines {
		if item == nil {
			continue
		}

		hostName := strings.TrimSpace(item.GetHost().GetHostName())
		machineUUID := strings.TrimSpace(item.GetHost().GetMachineUuid())

		// Table display values (used for auto-merge).
		endpointCell := endpoint
		hostNameCell := hostName
		if machineUUID != "" {
			switch {
			case hostNameCell == "":
				hostNameCell = machineUUID
			default:
				hostNameCell = hostNameCell + "\n" + machineUUID
			}
		}
		if hostNameCell == "" {
			hostNameCell = "-"
		}

		// Backward compatibility: keep original json/yaml semantics.
		endpointValue := endpoint
		if machineUUID != "" {
			endpointValue = endpoint + "\n" + machineUUID
		}
		hostNameValue := hostName
		if hostNameValue == "" {
			hostNameValue = "-"
		}

		cpuName := ""
		if len(item.GetCpu().GetProcessors()) > 0 {
			cpuName = item.GetCpu().GetProcessors()[0].GetModel()
		}

		var diskTotal, diskAvail uint64
		for _, disk := range item.GetDisks() {
			if disk == nil {
				continue
			}
			diskTotal += disk.GetSizeBytes()
			for _, mount := range disk.GetMounts() {
				if mount == nil {
					continue
				}
				diskAvail += mount.GetFreeBytes()
			}
		}

		rows = append(rows, machineRow{
			Endpoint:      endpointValue,
			HostName:      hostNameValue,
			endpointCell:  endpointCell,
			hostNameCell:  hostNameCell,
			LocalIP:       item.GetNetwork().GetLocalIp(),
			CPUName:       cpuName,
			CPUCores:      item.GetCpu().GetTotalCores(),
			MemoryTotal:   formatBytes(item.GetMemory().GetTotalPhysicalBytes()),
			MemoryAvail:   formatBytes(item.GetMemory().GetAvailableBytes()),
			DiskTotal:     formatBytes(diskTotal),
			DiskAvailable: formatBytes(diskAvail),
			NetworkRx:     formatBytes(item.GetNetwork().GetTotalRxBytes()),
			NetworkTx:     formatBytes(item.GetNetwork().GetTotalTxBytes()),
		})
	}
	return rows
}

func renderRows(format string, rows []machineRow) error {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "json":
		b, err := json.MarshalIndent(rows, "", "  ")
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(append(b, '\n'))
		return nil
	case "yaml":
		b, err := yaml.Marshal(rows)
		if err != nil {
			return err
		}
		_, _ = os.Stdout.Write(b)
		return nil
	default:
		table := tablewriter.NewWriter(os.Stdout)
		// Merge identical values in ENDPOINT column to get a "rowspan" look.
		table.SetAutoMergeCellsByColumnIndex([]int{0})
		// Draw row separator lines so merged cells look visually merged.
		table.SetRowLine(true)
		table.SetHeader([]string{
			"ENDPOINT",
			"HOSTNAME",
			"IP",
			"CPU",
			"CORES",
			"MEMORY(T/A)",
			"DISK(T/A)",
			"NETWORK(RX/TX)",
		})
		for _, row := range rows {
			table.Append([]string{
				row.endpointCell,
				row.hostNameCell,
				row.LocalIP,
				row.CPUName,
				fmt.Sprintf("%d", row.CPUCores),
				row.MemoryTotal + " / " + row.MemoryAvail,
				row.DiskTotal + " / " + row.DiskAvailable,
				row.NetworkRx + " / " + row.NetworkTx,
			})
		}
		table.Render()
		return nil
	}
}

func formatBytes(v uint64) string {
	const unit = 1024
	if v < unit {
		return fmt.Sprintf("%dB", v)
	}
	div, exp := uint64(unit), 0
	for n := v / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f%ciB", float64(v)/float64(div), "KMGTPE"[exp])
}
