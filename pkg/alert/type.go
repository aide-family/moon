package alert

import "time"

type (
	KV map[string]string

	Labels            KV
	GroupLabels       KV
	Annotations       KV
	CommonLabels      KV
	CommonAnnotations KV

	Alert struct {
		Status       string      `json:"status"`
		Labels       Labels      `json:"labels"`
		Annotations  Annotations `json:"annotations"`
		StartsAt     time.Time   `json:"startsAt"`
		EndsAt       time.Time   `json:"endsAt"`
		GeneratorURL string      `json:"generatorURL"`
		Fingerprint  string      `json:"fingerprint"`
	}

	Data struct {
		Receiver          string            `json:"receiver"`
		Status            string            `json:"status"`
		Alerts            []Alert           `json:"alerts"`
		GroupLabels       GroupLabels       `json:"groupLabels"`
		CommonLabels      CommonLabels      `json:"commonLabels"`
		CommonAnnotations CommonAnnotations `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		TruncatedAlerts   int32             `json:"truncatedAlerts"`
	}
)
