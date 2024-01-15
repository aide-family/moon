package strategy

import (
	"testing"
)

func TestNewAlarm(t *testing.T) {
	// 定义一个 map[string]string 类型的数据结构
	data := map[string]any{
		"labels": map[string]any{
			"instance": "192.168.1.100",
		},
		"value": 100,
	}

	str := "{{ .labels.instance }} $$的值大于 {{$value }}"

	t.Log(Formatter(str, data))
}
