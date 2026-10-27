package data

import (
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// TestRuleVariablesStorage 校验变量的存储格式与设计文档一致，并且历史规则或
// 损坏的列值都按没有变量读出，读取路径不因为这一列失败。
func TestRuleVariablesStorage(t *testing.T) {
	items := []bo.AlertRuleVariable{
		{Name: "disk_free", Expr: `node_filesystem_avail_bytes / 1024 / 1024 / 1024`, Description: "磁盘剩余容量（GiB）"},
		{Name: "load", Expr: "1 + 1"},
	}
	raw := encodeRuleVariables(items)
	want := `[{"name":"disk_free","expr":"node_filesystem_avail_bytes / 1024 / 1024 / 1024","description":"磁盘剩余容量（GiB）"},{"name":"load","expr":"1 + 1","description":""}]`
	if raw != want {
		t.Fatalf("storage format = %s, want %s", raw, want)
	}
	decoded := decodeRuleVariables(raw)
	if len(decoded) != len(items) {
		t.Fatalf("decoded = %+v", decoded)
	}
	for i := range items {
		if decoded[i] != items[i] {
			t.Fatalf("decoded[%d] = %+v, want %+v", i, decoded[i], items[i])
		}
	}

	if got := encodeRuleVariables(nil); got != "[]" {
		t.Fatalf("empty variables encoded as %q, want []", got)
	}
	for _, raw := range []string{"", "not json", "null", "{}"} {
		if got := decodeRuleVariables(raw); len(got) != 0 {
			t.Fatalf("decode(%q) = %+v, want empty", raw, got)
		}
	}
}
