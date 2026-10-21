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
	// 实时表格的 12 个内置字段都必须能保存，包括规则名 rule。
	allFields := []bo.AlertColumn{
		{Key: "summary", Type: "FIELD"}, {Key: "datasource", Type: "FIELD"},
		{Key: "level", Type: "FIELD"}, {Key: "sample_value", Type: "FIELD"},
		{Key: "starts_at", Type: "FIELD"}, {Key: "duration", Type: "FIELD"},
		{Key: "intervened", Type: "FIELD"}, {Key: "status", Type: "FIELD"},
		{Key: "description", Type: "FIELD"}, {Key: "expr", Type: "FIELD"},
		{Key: "fingerprint", Type: "FIELD"}, {Key: "rule", Type: "FIELD"},
	}
	if got, err := normalizeColumns(allFields); err != nil || len(got) != len(allFields) {
		t.Fatalf("all fields = %d, err = %v", len(got), err)
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
