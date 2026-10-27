package service

import (
	"encoding/json"
	"net/url"
	"sort"
	"testing"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/go-kratos/kratos/v3/encoding/form"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// 分析接口的查询参数名与前端冻结的 camelCase 一致：HTTP 层 BindQuery 走 form codec，
// 对 proto 消息按 JSON 名解析，这里用同一个解码器覆盖全部公共参数。
func TestAlertAnalysisRequestQueryBinding(t *testing.T) {
	common := url.Values{
		"from":       {"2026-10-01"},
		"to":         {"2026-10-27"},
		"timeZone":   {"Asia/Shanghai"},
		"status":     {"SUPPRESSING"},
		"focusStart": {"2026-10-05T00:00:00Z"},
		"focusEnd":   {"2026-10-06T00:00:00Z"},
	}
	var summary v1.GetAlertAnalysisSummaryRequest
	if err := form.DecodeValues(&summary, common); err != nil {
		t.Fatalf("decode summary query: %v", err)
	}
	if summary.GetFrom() != "2026-10-01" || summary.GetTo() != "2026-10-27" ||
		summary.GetTimeZone() != "Asia/Shanghai" || summary.GetStatus() != "SUPPRESSING" ||
		summary.GetFocusStart() != "2026-10-05T00:00:00Z" || summary.GetFocusEnd() != "2026-10-06T00:00:00Z" {
		t.Fatalf("summary query = %+v", &summary)
	}

	for name, target := range map[string]proto.Message{
		"status":       &v1.GetAlertAnalysisStatusRequest{},
		"level":        &v1.GetAlertAnalysisLevelRequest{},
		"recovery":     &v1.GetAlertAnalysisRecoveryRequest{},
		"fingerprints": &v1.GetAlertAnalysisFingerprintsRequest{},
	} {
		if err := form.DecodeValues(target, common); err != nil {
			t.Fatalf("decode %s query: %v", name, err)
		}
	}

	var trend v1.GetAlertAnalysisTrendRequest
	if err := form.DecodeValues(&trend, url.Values{
		"from": {"2026-10-01"}, "to": {"2026-10-27"}, "timeZone": {"UTC"},
		"status": {"RECOVERED"}, "granularity": {"week"},
	}); err != nil {
		t.Fatalf("decode trend query: %v", err)
	}
	if trend.GetGranularity() != "week" || trend.GetStatus() != "RECOVERED" || trend.GetTimeZone() != "UTC" {
		t.Fatalf("trend query = %+v", &trend)
	}

	var fingerprints v1.GetAlertAnalysisFingerprintsRequest
	if err := form.DecodeValues(&fingerprints, url.Values{
		"from": {"2026-10-01"}, "to": {"2026-10-27"}, "limit": {"7"},
	}); err != nil {
		t.Fatalf("decode fingerprints query: %v", err)
	}
	if fingerprints.GetLimit() != 7 {
		t.Fatalf("fingerprints limit = %d, want 7", fingerprints.GetLimit())
	}
}

// 响应字段名是前端写死的契约，这里按 HTTP 层同一个 protojson 配置检查 JSON 键。
func TestAlertAnalysisResponseJSONNames(t *testing.T) {
	keys := func(message proto.Message) (map[string]json.RawMessage, []string) {
		t.Helper()
		payload, err := protojson.MarshalOptions{EmitUnpopulated: true}.Marshal(message)
		if err != nil {
			t.Fatalf("marshal: %v", err)
		}
		var decoded map[string]json.RawMessage
		if err := json.Unmarshal(payload, &decoded); err != nil {
			t.Fatalf("unmarshal %s: %v", payload, err)
		}
		got := make([]string, 0, len(decoded))
		for key := range decoded {
			got = append(got, key)
		}
		sort.Strings(got)
		return decoded, got
	}
	assertKeys := func(message proto.Message, want ...string) map[string]json.RawMessage {
		t.Helper()
		decoded, got := keys(message)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Fatalf("keys = %v, want %v", got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("keys = %v, want %v", got, want)
			}
		}
		return decoded
	}
	nestedKeys := func(raw json.RawMessage, want ...string) {
		t.Helper()
		var object map[string]json.RawMessage
		if err := json.Unmarshal(raw, &object); err != nil {
			t.Fatalf("unmarshal nested object: %v", err)
		}
		got := make([]string, 0, len(object))
		for key := range object {
			got = append(got, key)
		}
		sort.Strings(got)
		sort.Strings(want)
		if len(got) != len(want) {
			t.Fatalf("nested keys = %v, want %v", got, want)
		}
		for index := range want {
			if got[index] != want[index] {
				t.Fatalf("nested keys = %v, want %v", got, want)
			}
		}
	}
	firstItem := func(raw json.RawMessage) json.RawMessage {
		t.Helper()
		var items []json.RawMessage
		if err := json.Unmarshal(raw, &items); err != nil || len(items) != 1 {
			t.Fatalf("items = %s, err = %v", raw, err)
		}
		return items[0]
	}

	// 1) summary：KPI 行。
	assertKeys(&v1.GetAlertAnalysisSummaryResponse{},
		"from", "to", "timeZone", "complete", "events", "uniqueFingerprints", "repeated")

	// 2) trend：只含非空桶，桶字段为 start/events。
	trend := assertKeys(&v1.GetAlertAnalysisTrendResponse{
		Buckets: []*v1.AlertAnalysisBucket{{Start: "2026-10-05T00:00:00+08:00"}},
	}, "from", "to", "timeZone", "complete", "buckets")
	nestedKeys(firstItem(trend["buckets"]), "start", "events")

	// 3) status 分布。
	status := assertKeys(&v1.GetAlertAnalysisStatusResponse{
		Items: []*v1.AlertAnalysisStatusItem{{Status: "RECOVERED"}},
	}, "complete", "total", "items")
	nestedKeys(firstItem(status["items"]), "status", "events")

	// 4) level 分布。
	level := assertKeys(&v1.GetAlertAnalysisLevelResponse{
		Items: []*v1.AlertAnalysisLevelItem{{LevelUid: "lvl"}},
	}, "complete", "total", "items")
	nestedKeys(firstItem(level["items"]), "levelUid", "events")

	// 5) 重复指纹。
	fingerprints := assertKeys(&v1.GetAlertAnalysisFingerprintsResponse{
		Items: []*v1.AlertAnalysisFingerprintItem{{Fingerprint: "fp", RuleUid: "rule", RuleName: "rule"}},
	}, "complete", "items")
	nestedKeys(firstItem(fingerprints["items"]), "fingerprint", "ruleUid", "ruleName", "events")

	// 6) 恢复时长。
	recovery := assertKeys(&v1.GetAlertAnalysisRecoveryResponse{
		Auto: &v1.AlertAnalysisRecoveryStats{}, Manual: &v1.AlertAnalysisRecoveryStats{},
	}, "complete", "auto", "manual")
	nestedKeys(recovery["auto"], "medianSeconds", "p90Seconds", "samples")
	nestedKeys(recovery["manual"], "medianSeconds", "p90Seconds", "samples")
}
