package cluster

type Code int

const (
	// Disabled Indicates that the cluster is disabled.
	Disabled Code = iota
	// Stopped Indicates that the cluster is stopped.
	Stopped
	// Started Indicates that the cluster is started.
	Started
	// Waiting Indicates that waiting the cluster resource ready.
	Waiting
	// Ready Indicates that the cluster resource is already ready.
	Ready
)

var codes = []string{"disabled", "stopped", "started", "waiting", "ready"}

func (c Code) String() string {
	if int(c) < len(codes)-1 {
		return codes[c]
	}
	return "unknown"
}
