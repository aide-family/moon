package do

import (
	"encoding/json"
	"time"

	"github.com/moon-monitor/moon/pkg/api/common"
	"github.com/moon-monitor/moon/pkg/plugin/cache"
	"github.com/moon-monitor/moon/pkg/util/hash"
	"github.com/moon-monitor/moon/pkg/util/kv"
	"github.com/moon-monitor/moon/pkg/util/kv/label"
	"github.com/moon-monitor/moon/pkg/util/timex"
)

var _ cache.Object = (*Alert)(nil)

type Alert struct {
	Status       common.AlertStatus `json:"status,omitempty"`
	Labels       *label.Label       `json:"labels,omitempty"`
	Annotations  *label.Annotation  `json:"annotations,omitempty"`
	StartsAt     *time.Time         `json:"startsAt,omitempty"`
	EndsAt       *time.Time         `json:"endsAt,omitempty"`
	GeneratorURL string             `json:"generatorURL,omitempty"`
	Fingerprint  string             `json:"fingerprint,omitempty"`
	Value        float64            `json:"value,omitempty"`

	Duration    time.Duration `json:"duration,omitempty"`
	LastUpdated time.Time     `json:"lastUpdated,omitempty"`
}

func (a *Alert) MarshalBinary() (data []byte, err error) {
	return json.Marshal(a)
}

func (a *Alert) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, a)
}

func (a *Alert) UniqueKey() string {
	return a.GetFingerprint()
}

func (a *Alert) GetStatus() common.AlertStatus {
	return a.Status
}

func (a *Alert) GetLabels() *label.Label {
	return a.Labels
}

func (a *Alert) GetAnnotations() *label.Annotation {
	return a.Annotations
}

func (a *Alert) GetStartsAt() *time.Time {
	return a.StartsAt
}

func (a *Alert) GetEndsAt() *time.Time {
	return a.EndsAt
}

func (a *Alert) GetGeneratorURL() string {
	return a.GeneratorURL
}

func (a *Alert) GetFingerprint() string {
	if a.Fingerprint == "" {
		stringMap := kv.NewStringMap(a.Labels.ToMap())
		a.Fingerprint = hash.MD5(kv.SortString(stringMap))
	}
	return a.Fingerprint
}

func (a *Alert) Resolved() {
	a.Status = common.AlertStatus_resolved
	a.LastUpdated = timex.Now()
	a.EndsAt = &a.LastUpdated
}

func (a *Alert) Firing() {
	a.Status = common.AlertStatus_firing
	a.LastUpdated = timex.Now()
}

func (a *Alert) GetValue() float64 {
	return a.Value
}

func (a *Alert) GetDuration() time.Duration {
	return a.Duration
}

func (a *Alert) GetLastUpdated() time.Time {
	return a.LastUpdated
}

func (a *Alert) IsPending() bool {
	return a.Status == common.AlertStatus_pending
}

func (a *Alert) IsFiring() bool {
	return a.Status == common.AlertStatus_firing
}

func (a *Alert) IsResolved() bool {
	return a.Status == common.AlertStatus_resolved
}
