package alertrule

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/alerttmpl"
	"github.com/aide-family/Alioth/pkg/merr"
)

// TestChangedLevelUIDs 校验阈值调整也会被视为级别变化。否则旧 FIRING 事件
// 会一直保留旧条件算出的持续时间和采样值，用户看不到配置修改后的新事件。
func TestChangedLevelUIDs(t *testing.T) {
	previous := []bo.AlertRuleLevel{{LevelUID: "p1", Condition: "GT", Sample: "FOR", Threshold: "1", ForSeconds: 30}}
	next := []bo.AlertRuleLevel{{LevelUID: "p1", Condition: "GT", Sample: "FOR", Threshold: "2", ForSeconds: 30}}
	recoverAll, uids := changedLevelUIDs(previous, next)
	if recoverAll || len(uids) != 1 || uids[0] != "p1" {
		t.Fatalf("threshold change: recoverAll=%v uids=%v", recoverAll, uids)
	}

	recoverAll, uids = changedLevelUIDs(previous, previous)
	if recoverAll || len(uids) != 0 {
		t.Fatalf("unchanged: recoverAll=%v uids=%v", recoverAll, uids)
	}

	recoverAll, uids = changedLevelUIDs(previous, nil)
	if !recoverAll || len(uids) != 0 {
		t.Fatalf("removed all: recoverAll=%v uids=%v", recoverAll, uids)
	}
}

// TestValidateRuleTemplates 覆盖模板字面量里多余的转义引号：模板文本不是
// JSON/YAML 字符串，\" 会被 Go 模板解析器判为非法操作数，保存时就该拒绝。
func TestValidateRuleTemplates(t *testing.T) {
	valid := `当前值 {{ $value | printf "%.1f" }}%`
	if err := validateRuleTemplates(&bo.AlertRule{Summary: valid, Description: valid}); err != nil {
		t.Fatalf("valid template rejected: %v", err)
	}
	rendered, err := alerttmpl.Render(
		context.Background(), valid, alerttmpl.NewData(nil, "{}", 15.47, "", nil),
		"description", time.Now(), nil, nil,
	)
	if err != nil || rendered != "当前值 15.5%" {
		t.Fatalf("render = %q err = %v", rendered, err)
	}

	escaped := `当前值 {{ $value | printf \"%.1f\" }}%`
	if err := validateRuleTemplates(&bo.AlertRule{Description: escaped}); err == nil {
		t.Fatal("description with backslash-escaped quotes must be rejected")
	}
	if err := validateRuleTemplates(&bo.AlertRule{Summary: escaped}); err == nil {
		t.Fatal("summary with backslash-escaped quotes must be rejected")
	}
}

// TestValidateRuleVariables 覆盖自定义查询变量的数量、名字格式、保留名、
// 唯一性、长度和变量语句模板校验。设计约定变量语句只能用基础字段，因此
// 引用其它变量也应当拒绝。
func TestValidateRuleVariables(t *testing.T) {
	valid := &bo.AlertRule{Variables: []bo.AlertRuleVariable{
		{Name: "disk_free", Expr: `node_filesystem_avail_bytes{instance="{{ $labels.instance }}"} / 1024 / 1024 / 1024`, Description: "磁盘剩余容量（GiB）"},
		{Name: "_tmp2", Expr: "1 + 1"},
	}}
	if err := validateRuleVariables(valid); err != nil {
		t.Fatalf("valid variables rejected: %v", err)
	}

	tooMany := make([]bo.AlertRuleVariable, 0, maxRuleVariables+1)
	for i := 0; i <= maxRuleVariables; i++ {
		tooMany = append(tooMany, bo.AlertRuleVariable{Name: fmt.Sprintf("v%d", i), Expr: "1"})
	}

	cases := []struct {
		name string
		rule *bo.AlertRule
	}{
		{"too many", &bo.AlertRule{Variables: tooMany}},
		{"dash in name", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk-free", Expr: "1"}}}},
		{"chinese name", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "磁盘剩余", Expr: "1"}}}},
		{"digit leading", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "1disk", Expr: "1"}}}},
		{"name too long", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: strings.Repeat("a", 33), Expr: "1"}}}},
		{"reserved labels", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "labels", Expr: "1"}}}},
		{"reserved value", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "value", Expr: "1"}}}},
		{"reserved externalURL", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "externalURL", Expr: "1"}}}},
		{"reserved externalLabels", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "externalLabels", Expr: "1"}}}},
		{"duplicate", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: "1"}, {Name: "disk_free", Expr: "2"}}}},
		{"empty expr", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk_free"}}}},
		{"blank expr", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: "  \t"}}}},
		{"expr too long", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: strings.Repeat("1", maxVariableExprLen+1)}}}},
		{"description too long", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: "1", Description: strings.Repeat("说明", maxVariableDescLen)}}}},
		{"broken template", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: "{{ $labels.instance "}}}},
		{"cross variable", &bo.AlertRule{Variables: []bo.AlertRuleVariable{{Name: "disk_free", Expr: "{{ $other_var }}"}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateRuleVariables(tc.rule)
			if err == nil {
				t.Fatalf("invalid variables accepted: %+v", tc.rule.Variables)
			}
			if !merr.IsInvalid(err) {
				t.Fatalf("error is not a client invalid error: %v", err)
			}
		})
	}
}

// TestValidateRuleTemplatesWithVars 校验摘要和描述可以用已声明的变量名，
// 未声明时仍按未定义变量拒绝；变量表达式的求值结果不参与模板语法检查。
func TestValidateRuleTemplatesWithVars(t *testing.T) {
	declared := []bo.AlertRuleVariable{{Name: "disk_free", Expr: "1"}}
	rule := func(variables []bo.AlertRuleVariable) *bo.AlertRule {
		return &bo.AlertRule{
			Summary:     `剩余 {{ $disk_free | printf "%.1f" }} GiB`,
			Description: `{{ $labels.instance }} 磁盘不足，剩余 {{ $disk_free }} GiB`,
			Variables:   variables,
		}
	}
	if err := validateRuleTemplates(rule(declared)); err != nil {
		t.Fatalf("declared variable rejected: %v", err)
	}
	if err := validateRuleTemplates(rule(nil)); err == nil {
		t.Fatal("undeclared variable must be rejected")
	}
	// 名字不合法时不应被注入模板，仍按未定义变量拒绝。
	invalid := []bo.AlertRuleVariable{{Name: "disk-free", Expr: "1"}}
	if err := validateRuleTemplates(rule(invalid)); err == nil {
		t.Fatal("invalid variable name must not be injected into the template")
	}
}

// TestPrepareValidateRuleTemplatesWithoutVars 旧规则没有变量时行为不变：摘要只认基础字段。
func TestPrepareValidateRuleTemplatesWithoutVars(t *testing.T) {
	if err := validateRuleTemplates(&bo.AlertRule{Summary: `当前 {{ $value | printf "%.1f" }}%`}); err != nil {
		t.Fatalf("rule without variables rejected: %v", err)
	}
	if err := validateRuleVariables(&bo.AlertRule{}); err != nil {
		t.Fatalf("rule without variables rejected: %v", err)
	}
}
