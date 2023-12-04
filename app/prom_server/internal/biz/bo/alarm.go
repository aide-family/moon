package bo

import (
	"encoding/json"

	"prometheus-manager/pkg/strategy"
)

type (
	GroupLabels map[string]string

	CommonLabels map[string]string

	CommonAnnotations map[string]string

	AlarmBo struct {
		Receiver          string            `json:"receiver"`
		Status            string            `json:"status"`
		Alerts            []*AlertBo        `json:"alerts"`
		GroupLabels       GroupLabels       `json:"groupLabels"`
		CommonLabels      CommonLabels      `json:"commonLabels"`
		CommonAnnotations CommonAnnotations `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		TruncatedAlerts   int32             `json:"truncatedAlerts"`
	}

	AlertBo struct {
		Status       string                `json:"status"`
		Labels       *strategy.Labels      `json:"labels"`
		Annotations  *strategy.Annotations `json:"annotations"`
		StartsAt     int64                 `json:"startsAt"`
		EndsAt       int64                 `json:"endsAt"`
		GeneratorURL string                `json:"generatorURL"`
		Fingerprint  string                `json:"fingerprint"`
	}
)

// String .
func (b *AlertBo) String() string {
	bytes, _ := json.Marshal(b)
	return string(bytes)
}

// GetStatus .
func (b *AlertBo) GetStatus() string {
	return b.Status
}

// GetLabels .
func (b *AlertBo) GetLabels() *strategy.Labels {
	return b.Labels
}

// ToLabelsMap .
func (b *AlertBo) ToLabelsMap() map[string]string {
	if b.Labels == nil {
		return nil
	}
	return *b.Labels
}

// GetAnnotations .
func (b *AlertBo) GetAnnotations() *strategy.Annotations {
	return b.Annotations
}

// ToAnnotationsMap .
func (b *AlertBo) ToAnnotationsMap() map[string]string {
	if b.Annotations == nil {
		return nil
	}
	return *b.Annotations
}
