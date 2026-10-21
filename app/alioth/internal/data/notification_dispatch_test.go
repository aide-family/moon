package data

import (
	"encoding/json"
	"testing"
)

func TestLabelsMatch(t *testing.T) {
	tests := []struct {
		name     string
		selector string
		labels   string
		want     bool
	}{
		{name: "empty selector", selector: "{}", labels: `{"env":"prod"}`, want: true},
		{name: "exact match", selector: `{"env":"prod"}`, labels: `{"env":"prod","team":"ops"}`, want: true},
		{name: "regex match", selector: `{"instance":"10\\.0\\..*"}`, labels: `{"instance":"10.0.0.1:9100"}`, want: true},
		{name: "regex is fully anchored", selector: `{"env":"prod"}`, labels: `{"env":"production"}`, want: false},
		{name: "regex alternation", selector: `{"env":"prod|staging"}`, labels: `{"env":"staging"}`, want: true},
		{name: "missing key", selector: `{"env":"prod"}`, labels: `{"team":"ops"}`, want: false},
		{name: "empty value requires the key", selector: `{"env":""}`, labels: `{"team":"ops"}`, want: false},
		{name: "different value", selector: `{"env":"prod"}`, labels: `{"env":"dev"}`, want: false},
		{name: "invalid event labels", selector: `{"env":"prod"}`, labels: `not-json`, want: false},
		{name: "invalid pattern falls back to literal", selector: `{"env":"prod["}`, labels: `{"env":"prod["}`, want: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := labelsMatch(tt.selector, tt.labels); got != tt.want {
				t.Fatalf("labelsMatch(%q, %q) = %v, want %v", tt.selector, tt.labels, got, tt.want)
			}
		})
	}
}

func TestContainsString(t *testing.T) {
	if !containsString([]string{"a", "b"}, "b") {
		t.Fatal("expected b")
	}
	if containsString([]string{"a"}, "b") {
		t.Fatal("unexpected b")
	}
}

// 通知任务参数要和告警模板共用同一套字段，Prometheus 写法在投递时才能展开。
func TestParamsJSONCarriesTemplateFields(t *testing.T) {
	raw := paramsJSON(notifyEvent{
		EventUID: "evt-1", Phase: "FIRING", Summary: "测试告警", Description: "测试告警, value: 1",
		Labels: `{"job":"api"}`, Value: "1", ExternalURL: "http://prometheus.example", Fingerprint: "fp",
	})
	params := map[string]any{}
	if err := json.Unmarshal([]byte(raw), &params); err != nil {
		t.Fatalf("params json = %s: %v", raw, err)
	}
	if params["phase"] != "FIRING" || params["summary"] != "测试告警" || params["eventUid"] != "evt-1" {
		t.Fatalf("event fields = %#v", params)
	}
	if params["labels"] != `{"job":"api"}` {
		t.Fatalf("labels placeholder changed: %#v", params["labels"])
	}
	labels, ok := params["Labels"].(map[string]any)
	if !ok || labels["job"] != "api" {
		t.Fatalf("structured labels = %#v", params["Labels"])
	}
	if params["Value"] != 1.0 || params["value"] != 1.0 {
		t.Fatalf("sample value = %#v", params["Value"])
	}
	if params["ExternalURL"] != "http://prometheus.example" || params["externalURL"] != "http://prometheus.example" {
		t.Fatalf("external url = %#v", params["ExternalURL"])
	}
}
