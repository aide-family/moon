package bo

import "time"

// WorkerInstance describes one live or recently-seen background worker process.
type WorkerInstance struct {
	Service       string
	InstanceUID   string
	Hostname      string
	Busy          bool
	TaskUID       string
	TaskName      string
	TaskStartedAt *time.Time
	LastSeenAt    time.Time
	Online        bool
}

// WorkerStatus is the heartbeat payload accepted from an internal worker.
type WorkerStatus struct {
	Service       string
	InstanceUID   string
	Hostname      string
	Busy          bool
	TaskUID       string
	TaskName      string
	TaskStartedAt *time.Time
}
