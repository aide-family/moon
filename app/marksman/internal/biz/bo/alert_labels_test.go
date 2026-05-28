package bo

import (
	"testing"

	"github.com/aide-family/magicbox/strutil/cnst"
)

func TestIsReservedAlertSystemLabelKey(t *testing.T) {
	if !IsReservedAlertSystemLabelKey(cnst.LabelNamespaceUID) {
		t.Fatal("namespace_uid should be reserved")
	}
	if !IsReservedAlertSystemLabelKey(cnst.LabelStrategyUID) {
		t.Fatal("strategy_uid should be reserved")
	}
	if IsReservedAlertSystemLabelKey("team") {
		t.Fatal("custom label should not be reserved")
	}
}

func TestFilterLabelsForAlertFingerprintIgnoresVolatileNameLabels(t *testing.T) {
	labels := map[string]string{
		cnst.LabelNamespaceUID:      "1",
		cnst.LabelStrategyUID:       "3",
		cnst.LabelDatasourceUID:     "10",
		cnst.LabelAlertName:         "OldName",
		cnst.LabelStrategyGroupName: "group-a",
		cnst.LabelSeverity:          "critical",
		"instance":                  "host-a",
	}
	filtered := FilterLabelsForAlertFingerprint(labels)
	if filtered[cnst.LabelAlertName] != "" {
		t.Fatal("alertname should be excluded from fingerprint labels")
	}
	if filtered[cnst.LabelStrategyGroupName] != "" {
		t.Fatal("strategy_group_name should be excluded from fingerprint labels")
	}
	if filtered[cnst.LabelNamespaceUID] != "1" {
		t.Fatal("namespace_uid should remain in fingerprint labels")
	}
	if filtered["instance"] != "host-a" {
		t.Fatal("series labels should remain in fingerprint labels")
	}
}

func TestFilterLabelsForAlertFingerprintStableWhenNamesChange(t *testing.T) {
	base := map[string]string{
		cnst.LabelNamespaceUID:  "1",
		cnst.LabelStrategyUID:   "3",
		cnst.LabelDatasourceUID: "10",
		"instance":              "host-a",
	}
	a := mapsCloneString(base)
	a[cnst.LabelAlertName] = "OldName"
	a[cnst.LabelStrategyGroupName] = "group-a"
	b := mapsCloneString(base)
	b[cnst.LabelAlertName] = "NewName"
	b[cnst.LabelStrategyGroupName] = "group-b"
	filteredA := FilterLabelsForAlertFingerprint(a)
	filteredB := FilterLabelsForAlertFingerprint(b)
	if len(filteredA) != len(filteredB) {
		t.Fatalf("filtered label count mismatch: %d vs %d", len(filteredA), len(filteredB))
	}
	for key, value := range filteredA {
		if filteredB[key] != value {
			t.Fatalf("filtered labels differ at %q: %q vs %q", key, value, filteredB[key])
		}
	}
}

func mapsCloneString(m map[string]string) map[string]string {
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
