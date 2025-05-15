package bo

import (
	"time"

	"github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/plugin/server"
	"github.com/moon-monitor/moon/pkg/util/kv/label"
)

type Alert interface {
	GetStatus() common.AlertStatus
	GetLabels() *label.Label
	GetAnnotations() *label.Annotation
	GetStartsAt() *time.Time
	GetEndsAt() *time.Time
	GetGeneratorURL() string
	GetFingerprint() string
	GetValue() float64
	Resolved()
	Firing()
	IsResolved() bool
	IsFiring() bool
	IsPending() bool
	GetDuration() time.Duration
	GetLastUpdated() time.Time
}

type AlertJob interface {
	GetAlert() Alert
	server.CronJob
}
