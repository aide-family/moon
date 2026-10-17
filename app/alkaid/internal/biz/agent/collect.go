package agent

import (
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"
)

type machineCollect struct {
	OS         string `json:"os"`
	Arch       string `json:"arch"`
	OSVersion  string `json:"osVersion"`
	CPU        string `json:"cpu"`
	Disk       string `json:"disk"`
	Memory     string `json:"memory"`
	BootTime   string `json:"bootTime"`
	TimeServer string `json:"timeServer"`
}

func collectJSON() string {
	payload := machineCollect{
		OS: runtime.GOOS, Arch: runtime.GOARCH, CPU: collectCPU(),
		Disk: collectDisk(), Memory: collectMemory(), TimeServer: collectTimeServer(),
	}
	if info, err := host.Info(); err == nil {
		if info.OS != "" {
			payload.OS = info.OS
		}
		if info.KernelArch != "" {
			payload.Arch = info.KernelArch
		}
		payload.OSVersion = strings.TrimSpace(strings.TrimSpace(info.Platform) + " " + strings.TrimSpace(info.PlatformVersion))
		if info.BootTime > 0 {
			payload.BootTime = time.Unix(int64(info.BootTime), 0).UTC().Format(time.RFC3339)
		}
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return string(raw)
}

func collectCPU() string {
	count, err := cpu.Counts(true)
	if err != nil || count <= 0 {
		count = runtime.NumCPU()
	}
	text := strconv.Itoa(count)
	infos, err := cpu.Info()
	if err != nil || len(infos) == 0 {
		return text
	}
	model := strings.TrimSpace(infos[0].ModelName)
	if model == "" {
		return text
	}
	return text + " x " + model
}

func collectDisk() string {
	usage, err := disk.Usage("/")
	if err != nil {
		return ""
	}
	return formatBytes(usage.Used) + "/" + formatBytes(usage.Total)
}

func collectMemory() string {
	stat, err := mem.VirtualMemory()
	if err != nil {
		return ""
	}
	return formatBytes(stat.Total)
}

func collectTimeServer() string {
	switch runtime.GOOS {
	case "linux":
		if server := commandTimeServer("timedatectl", "show-timesync", "-p", "ServerName", "--value"); server != "" {
			return server
		}
		if server := firstNTPLine("/etc/systemd/timesyncd.conf", "NTP", "FallbackNTP"); server != "" {
			return server
		}
		return firstNTPLine("/etc/chrony/chrony.conf", "server", "pool")
	case "darwin":
		return commandTimeServer("systemsetup", "-getnetworktimeserver")
	default:
		return ""
	}
}

func commandTimeServer(name string, args ...string) string {
	out, err := exec.Command(name, args...).Output()
	if err != nil {
		return ""
	}
	text := strings.TrimSpace(string(out))
	if text == "" || strings.EqualFold(text, "n/a") {
		return ""
	}
	if index := strings.LastIndex(text, ":"); index >= 0 {
		text = strings.TrimSpace(text[index+1:])
	}
	return firstToken(text)
}

func firstNTPLine(path string, keys ...string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		key, value := fields[0], ""
		if len(fields) > 1 {
			value = fields[1]
		}
		if parsed, rest, ok := strings.Cut(key, "="); ok {
			key = parsed
			value = rest
		}
		if value == "" {
			continue
		}
		for _, want := range keys {
			if strings.EqualFold(key, want) {
				return firstToken(value)
			}
		}
	}
	return ""
}

func firstToken(value string) string {
	fields := strings.Fields(value)
	if len(fields) == 0 {
		return ""
	}
	return fields[0]
}

func formatBytes(n uint64) string {
	units := []string{"B", "KiB", "MiB", "GiB", "TiB"}
	value := float64(n)
	index := 0
	for value >= 1024 && index < len(units)-1 {
		value /= 1024
		index++
	}
	if index == 0 {
		return strconv.FormatUint(n, 10) + "B"
	}
	return strconv.FormatFloat(value, 'f', 1, 64) + units[index]
}
