package strategy

import (
	"testing"
)

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
