package alertrule

import (
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
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
