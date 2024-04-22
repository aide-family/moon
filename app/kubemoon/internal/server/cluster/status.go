package cluster

type Code int

const (
	// Disabled Indicates that the cluster is disabled.
	Disabled Code = iota
	// Offline Indicates that the cluster is offline and cannot work.
	Offline
	// Unhealthy Indicates that the cluster is online but the health check failed
	Unhealthy
	// Ready Indicates that the cluster is ready to work
	Ready
	// Stopped Indicates that the cluster is ready but in a stopped state
	Stopped
	// Running Indicates that the cluster is ready and working
	Running
)

var codes = []string{"disabled", "offline", "unhealthy", "ready", "stopped", "running"}

func (c Code) String() string {
	if int(c) >= 0 && int(c) < len(codes) {
		return codes[c]
	}
	return "unknown"
}
