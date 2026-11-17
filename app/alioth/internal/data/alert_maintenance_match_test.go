package data

import (
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

// TestWindowMatchesEventLabels 确认维护窗口的标签匹配语义：整串匹配、空 matcher
// 不按标签收窄，历史行里的非法正则按不匹配处理而不是 panic。
func TestWindowMatchesEventLabels(t *testing.T) {
	withLabels := func(matcher string) bo.AlertMaintenanceWindow {
		return bo.AlertMaintenanceWindow{Labels: matcher}
	}

	t.Run("matches anchored pattern", func(t *testing.T) {
		if !windowMatchesEvent(withLabels(`{"env":"prod"}`), `{"env":"prod"}`, "", "", "") {
			t.Fatal("prod should match prod")
		}
		if windowMatchesEvent(withLabels(`{"env":"prod"}`), `{"env":"prod-old"}`, "", "", "") {
			t.Fatal("prod must not match prod-old")
		}
	})

	t.Run("regex pattern", func(t *testing.T) {
		if !windowMatchesEvent(withLabels(`{"env":"prod-.*"}`), `{"env":"prod-2"}`, "", "", "") {
			t.Fatal("prod-.* should match prod-2")
		}
		if windowMatchesEvent(withLabels(`{"env":"prod-.*"}`), `{"env":"staging"}`, "", "", "") {
			t.Fatal("prod-.* must not match staging")
		}
	})

	t.Run("missing label does not match", func(t *testing.T) {
		if windowMatchesEvent(withLabels(`{"env":"prod"}`), `{"service":"db"}`, "", "", "") {
			t.Fatal("missing label must not match")
		}
	})

	t.Run("empty matcher does not narrow by labels", func(t *testing.T) {
		if !windowMatchesEvent(withLabels("{}"), `{"env":"staging"}`, "", "", "") {
			t.Fatal("empty matcher should match any labels")
		}
	})

	// 写入侧已拒绝非法正则，但校验之前入库的行仍在：匹配链路在调度和回写路径上，
	// 这里只能是「不匹配」，不能让 regexp 的 panic 带走进程。
	t.Run("invalid legacy pattern does not panic", func(t *testing.T) {
		if windowMatchesEvent(withLabels(`{"env":"("}`), `{"env":"prod"}`, "", "", "") {
			t.Fatal("invalid pattern must not match")
		}
	})

	t.Run("level filter combines with labels", func(t *testing.T) {
		window := bo.AlertMaintenanceWindow{Labels: `{"env":"prod"}`, LevelUIDs: []string{"level-1"}}
		if !windowMatchesEvent(window, `{"env":"prod"}`, "level-1", "", "") {
			t.Fatal("matching level and labels should match")
		}
		if windowMatchesEvent(window, `{"env":"prod"}`, "level-2", "", "") {
			t.Fatal("different level must not match")
		}
	})
}
