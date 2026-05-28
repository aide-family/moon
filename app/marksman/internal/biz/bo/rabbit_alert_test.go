package bo

import (
	"testing"

	"github.com/aide-family/magicbox/strutil/cnst"
)

func TestBuildRabbitAlertLabelsDoesNotDuplicateExistingKeys(t *testing.T) {
	extra := map[string]string{
		cnst.LabelNamespaceUID:  "1",
		cnst.LabelStrategyUID:   "3",
		cnst.LabelAlertName:     "HighCPU",
		cnst.LabelSeverity:      "critical",
		cnst.LabelDatasourceUID: "5",
		"instance":              "host-a",
	}
	labels := buildRabbitAlertLabels(
		100,
		1,
		"fp-1",
		2,
		"group-a",
		3,
		"HighCPU",
		4,
		"critical",
		5,
		"prom-main",
		"prod",
		extra,
	)
	if labels[cnst.LabelNamespaceUID] != "1" {
		t.Fatalf("namespace_uid = %q, want 1", labels[cnst.LabelNamespaceUID])
	}
	if labels[cnst.LabelAlertEventUID] != "100" {
		t.Fatalf("alert_event_uid = %q, want 100", labels[cnst.LabelAlertEventUID])
	}
	if labels["instance"] != "host-a" {
		t.Fatal("custom label should be preserved")
	}
}
