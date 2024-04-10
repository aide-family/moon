package bo

import (
	"encoding/json"

	"github.com/aide-family/moon/pkg/strategy"
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
		StartsAt     string                `json:"startsAt"`
		EndsAt       string                `json:"endsAt"`
		GeneratorURL string                `json:"generatorURL"`
		Fingerprint  string                `json:"fingerprint"`
		Value        float64               `json:"value"`
	}
)

// Bytes .
func (b *AlertBo) Bytes() []byte {
	bytes, _ := json.Marshal(b)
	return bytes
}

// String .
func (b *AlertBo) String() string {
	return string(b.Bytes())
}

// ToMap .
func (b *AlertBo) ToMap() map[string]any {
	m := make(map[string]any)
	_ = json.Unmarshal([]byte(b.String()), &m)
	return m
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

// StringToAlertBo .
func StringToAlertBo(str string) *AlertBo {
	var alertBo AlertBo
	_ = json.Unmarshal([]byte(str), &alertBo)
	return &alertBo
}
