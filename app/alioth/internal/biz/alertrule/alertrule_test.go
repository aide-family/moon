package alertrule

import (
	"context"
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
	"github.com/aide-family/Alioth/pkg/alerttmpl"
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
