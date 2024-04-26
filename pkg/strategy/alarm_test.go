package strategy

import (
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/util/hash"
)

type AlertBo struct {
	Status       string       `json:"status"`
	Labels       *Labels      `json:"labels"`
	Annotations  *Annotations `json:"annotations"`
	StartsAt     string       `json:"startsAt"`
	EndsAt       string       `json:"endsAt"`
	GeneratorURL string       `json:"generatorURL"`
	Fingerprint  string       `json:"fingerprint"`
	Value        float64      `json:"value"`
}

func TestNewAlarm(t *testing.T) {
	// 定义一个 map[string]string 类型的数据结构
	data := map[string]any{
		"labels": Labels{
			"instance": "192.168.1.100",
		},
		"annotations": Annotations{
			"summary": "这是一个测试告警, xxx,1234,55,ABC,aBc",
		},
		"value": 100.1234,
	}

	str := `{{ .labels.instance }} $$的值大于 {{$value }} v: {{ printf "%.2f" .value }} 
Current Time: {{ now.Format "2006-01-02 15:04:05" }}
Current Time: {{ now.Unix }}
Annotation {{ annotations.String }}
HasPrefix {{ hasPrefix .annotations.summary "这" }}
Split {{ split .annotations.summary "," }}
toUpper {{ toUpper .annotations.summary }}
toLower {{ toLower .annotations.summary }}
Label {{ labels.String }}`

	t.Log(Formatter(str, data))
}

func TestFormatter(t *testing.T) {
	labels := Labels{
		"__name__":       "test",
		"__alert_id__":   "1",
		"__group_name__": "test",
		"__group_id__":   "1",
		"__level_id__":   "1",
		"instance":       "test-instance",
		"alertname":      "test-alert",
		"ip":             "192.168.1.100",
	}
	annotations := Annotations{
		"summary":     "test summary",
		"description": "test description",
	}
	data := &AlertBo{
		Status:       "firing",
		Labels:       &labels,
		Annotations:  &annotations,
		StartsAt:     time.Now().Format(time.RFC3339),
		EndsAt:       time.Now().Add(time.Hour).Format(time.RFC3339),
		GeneratorURL: "https://prometheus.aide-cloud.cn/#/home",
		Fingerprint:  hash.MD5("test"),
		Value:        100.0011,
	}

	templateStr := `
告警状态: {{ .Status }}
告警标签: {{ .Labels }}
	机器实例: {{ .Labels.instance }}
	告警规则名称: {{ .Labels.alertname }}
告警内容: {{ .Annotations }}
	告警描述: {{ .Annotations.summary }}
	告警详情: {{ .Annotations.description }}
告警时间: {{ .StartsAt }}
恢复时间: {{ .EndsAt }}
链接地址: {{ .GeneratorURL }}
告警指纹: {{ .Fingerprint }}
当前值: {{ .Value }}
当前时间: {{ now.Format "2006-01-02 15:04:05" }}
是否告警: {{ if contains .Status "firing" }}告警了{{ else }}恢复了{{ end }}
IP:{{ range split .Labels.ip "." }}
 an ip {{ . }}
{{- end }}
`
	t.Log(Formatter(templateStr, data))
}
