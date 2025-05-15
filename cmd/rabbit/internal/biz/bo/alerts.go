package bo

import (
	"strconv"
	"strings"

	"github.com/moon-monitor/moon/pkg/util/kv/label"
	"github.com/moon-monitor/moon/pkg/util/validate"
)

type AlertItem struct {
	Status       string
	Labels       *label.Label
	Annotations  *label.Annotation
	StartsAt     string
	EndsAt       string
	GeneratorURL string
	Fingerprint  string
	Value        string
}

type AlertsItem struct {
	Receiver          string
	Status            string
	Alerts            []*AlertItem
	GroupLabels       *label.Label
	CommonLabels      *label.Label
	CommonAnnotations *label.Annotation
	ExternalURL       string
	Version           string
	GroupKey          string
	TruncatedAlerts   int32
}

// GetReceiver implements bo.AlertsItem.
func (a *AlertsItem) GetReceiver() []string {
	if a == nil || validate.TextIsNull(a.Receiver) {
		return []string{}
	}
	return strings.Split(a.Receiver, ",")
}

func (a *AlertsItem) GetTeamID() string {
	if a == nil {
		return ""
	}
	return strconv.Itoa(int(a.CommonLabels.GetTeamId()))
}
