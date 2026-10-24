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

// 历史告警明细的列配置可以保存 ends_at：历史事件已经结束，结束时间是台账列；
// 实时视图不提供这一列，但白名单是视图无关的，只要字段本身合法就允许保存。
func TestNormalizeHistoryOnlyColumns(t *testing.T) {
	columns, err := normalizeColumns([]bo.AlertColumn{
		{Key: "starts_at", Type: "FIELD"},
		{Key: "ends_at", Type: "FIELD"},
		{Key: "duration", Type: "FIELD"},
	})
	if err != nil {
		t.Fatalf("ends_at should be accepted: %v", err)
	}
	if len(columns) != 3 || columns[1].Key != "ends_at" {
		t.Fatalf("columns = %+v", columns)
	}
}

func TestBuiltinColumnScopes(t *testing.T) {
	// 内置视图不需要告警页面就能保存列配置；空 scope 仍然必须被拒绝。
	for _, scope := range []string{tabMine, tabHistory} {
		if _, ok := builtinColumnScopes[scope]; !ok {
			t.Fatalf("scope %q should be builtin", scope)
		}
	}
	if _, ok := builtinColumnScopes[""]; ok {
		t.Fatal("empty scope must not be builtin")
	}
	if _, ok := builtinColumnScopes["page-uid"]; ok {
		t.Fatal("page uid must be validated against the alert page, not treated as builtin")
	}
}
