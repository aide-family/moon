package alert

import "encoding/json"

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
		StartsAt     int64       `json:"startsAt"`
		EndsAt       int64       `json:"endsAt"`
		GeneratorURL string      `json:"generatorURL"`
		Fingerprint  string      `json:"fingerprint"`
	}

	Data struct {
		Receiver          string            `json:"receiver"`
		Status            string            `json:"status"`
		Alerts            []*Alert          `json:"alerts"`
		GroupLabels       GroupLabels       `json:"groupLabels"`
		CommonLabels      CommonLabels      `json:"commonLabels"`
		CommonAnnotations CommonAnnotations `json:"commonAnnotations"`
		ExternalURL       string            `json:"externalURL"`
		Version           string            `json:"version"`
		GroupKey          string            `json:"groupKey"`
		TruncatedAlerts   int32             `json:"truncatedAlerts"`
	}
)

func (l *Data) Byte() []byte {
	if l == nil {
		return nil
	}
	b, _ := json.Marshal(*l)
	return b
}

func (l KV) String() string {
	if l == nil {
		return ""
	}

	str, _ := json.Marshal(l)
	return string(str)
}

func ToLabels(str string) Labels {
	var labels Labels
	_ = json.Unmarshal([]byte(str), &labels)
	return labels
}

func ToAnnotations(str string) Annotations {
	var annotations Annotations
	_ = json.Unmarshal([]byte(str), &annotations)
	return annotations
}
