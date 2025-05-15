package bo

import (
	"time"

	"github.com/aide-family/moon/pkg/api/common"
	"github.com/aide-family/moon/pkg/plugin/server"
	"github.com/aide-family/moon/pkg/util/kv/label"
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
