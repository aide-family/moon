package bo

import (
	"github.com/aide-family/moon/pkg/util/types"
	"github.com/aide-family/moon/pkg/vobj"
	"github.com/aide-family/moon/pkg/watch"
)

var _ watch.Indexer = (*Alarm)(nil)
var _ watch.Indexer = (*Alert)(nil)

type (
	// Alarm alarm detail info
	Alarm struct {
		Receiver          string           `json:"receiver"`
		Status            vobj.AlertStatus `json:"status"`
		Alerts            []*Alert         `json:"alerts"`
		GroupLabels       *vobj.Labels     `json:"groupLabels"`
		CommonLabels      *vobj.Labels     `json:"commonLabels"`
		CommonAnnotations vobj.Annotations `json:"commonAnnotations"`
		ExternalURL       string           `json:"externalURL"`
		Version           string           `json:"version"`
		GroupKey          string           `json:"groupKey"`
		TruncatedAlerts   int32            `json:"truncatedAlerts"`
	}

	// Alert alert detail info
	Alert struct {
		Status       vobj.AlertStatus `json:"status"`
		Labels       *vobj.Labels     `json:"labels"`
		Annotations  vobj.Annotations `json:"annotations"`
		StartsAt     *types.Time      `json:"startsAt"`
		EndsAt       *types.Time      `json:"endsAt"`
		GeneratorURL string           `json:"generatorURL"`
		Fingerprint  string           `json:"fingerprint"`
		Value        float64          `json:"value"`
	}
)

// Index gen alert index
func (a *Alert) Index() string {
	//TODO implement me
	panic("implement me")
}

// Index gen alarm index
func (a *Alarm) Index() string {
	//TODO implement me
	panic("implement me")
}

// Message gen alarm message
func (a *Alarm) Message() *watch.Message {
	return watch.NewMessage(a, vobj.TopicAlarm)
}

// Message gen alert message
func (a *Alert) Message() *watch.Message {
	return watch.NewMessage(a, vobj.TopicAlert)
}
