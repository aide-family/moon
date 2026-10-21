package alertpage

import (
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func TestNormalizeColumns(t *testing.T) {
	columns, err := normalizeColumns([]bo.AlertColumn{
		{Key: "summary", Type: "field"},
		{Key: "node", Type: "label", Label: "节点"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(columns) != 2 || columns[0].Type != ColumnKindField || columns[1].Type != ColumnKindLabel || columns[1].Label != "节点" {
		t.Fatalf("columns = %+v", columns)
	}
	if _, err := normalizeColumns([]bo.AlertColumn{{Key: "no-such-field", Type: "FIELD"}}); err == nil {
		t.Fatal("unknown field should be rejected")
	}
	if _, err := normalizeColumns([]bo.AlertColumn{{Key: "node", Type: "LABEL"}, {Key: "node", Type: "LABEL"}}); err == nil {
		t.Fatal("duplicate columns should be rejected")
	}
	if _, err := normalizeColumns([]bo.AlertColumn{{Key: "1node", Type: "LABEL"}}); err == nil {
		t.Fatal("invalid label key should be rejected")
	}
}
