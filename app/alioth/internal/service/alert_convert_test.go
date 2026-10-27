package service

import (
	"testing"

	v1 "github.com/aide-family/Alioth/api/alioth/v1"
	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

// TestRuleVariableConvert 校验自定义查询变量在 proto 与 bo 之间往返，顺序和
// 字段都保持不变；旧规则没有变量时返回空切片而不是 nil。
func TestRuleVariableConvert(t *testing.T) {
	inputs := []*v1.AlertRuleVariableInput{
		{Name: "disk_free", Expr: `node_filesystem_avail_bytes`, Description: "磁盘剩余容量"},
		{Name: "load", Expr: "1 + 1"},
	}
	variables := variableInputs(inputs)
	if len(variables) != len(inputs) {
		t.Fatalf("variableInputs = %+v", variables)
	}
	item := ruleItem(bo.AlertRule{Name: "磁盘", Variables: variables})
	if len(item.GetVariables()) != len(inputs) {
		t.Fatalf("ruleItem variables = %+v", item.GetVariables())
	}
	for i, input := range inputs {
		got := item.GetVariables()[i]
		if got.GetName() != input.GetName() || got.GetExpr() != input.GetExpr() || got.GetDescription() != input.GetDescription() {
			t.Fatalf("variable %d = %+v, want %+v", i, got, input)
		}
	}

	if got := ruleItem(bo.AlertRule{}).GetVariables(); len(got) != 0 {
		t.Fatalf("rule without variables = %+v, want empty", got)
	}
	if got := variableInputs(nil); len(got) != 0 {
		t.Fatalf("variableInputs(nil) = %+v, want empty", got)
	}
}

// TestSnapshotItemVariables 校验快照下发时变量按保存顺序传给 megrez。
func TestSnapshotItemVariables(t *testing.T) {
	item := snapshotItem(bo.RuleSnapshot{
		Rule: bo.AlertRule{
			Name: "磁盘", Summary: "剩余 {{ $disk_free }}",
			Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: "node_filesystem_avail_bytes", Description: "磁盘剩余容量"}},
		},
	})
	variables := item.GetVariables()
	if len(variables) != 1 || variables[0].GetName() != "disk_free" || variables[0].GetExpr() != "node_filesystem_avail_bytes" || variables[0].GetDescription() != "磁盘剩余容量" {
		t.Fatalf("snapshot variables = %+v", variables)
	}
	if got := snapshotItem(bo.RuleSnapshot{}).GetVariables(); len(got) != 0 {
		t.Fatalf("snapshot without variables = %+v, want empty", got)
	}
}
