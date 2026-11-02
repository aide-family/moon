package alertmaintenance

import (
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func validWindow() bo.AlertMaintenanceWindow {
	return bo.AlertMaintenanceWindow{
		Name:     "数据库维护",
		Labels:   `{"env":"prod","service":"db"}`,
		StartsAt: time.Date(2026, 11, 3, 1, 0, 0, 0, time.UTC),
		EndsAt:   time.Date(2026, 11, 3, 3, 0, 0, 0, time.UTC),
	}
}

func TestNormalizeWindow(t *testing.T) {
	in := validWindow()
	in.Name = "  数据库维护  "
	in.LevelUIDs = []string{"level-1", " level-1 ", "level-2", ""}
	in.DatasourceUIDs = []string{"ds-1", "ds-1"}
	in.RuleUIDs = []string{"rule-1", "rule-2"}
	if err := normalizeWindow(&in); err != nil {
		t.Fatalf("valid maintenance window rejected: %v", err)
	}
	if in.Name != "数据库维护" {
		t.Fatalf("name = %q, want trimmed", in.Name)
	}
	if in.Labels != `{"env":"prod","service":"db"}` {
		t.Fatalf("labels = %q", in.Labels)
	}
	if len(in.LevelUIDs) != 2 || len(in.DatasourceUIDs) != 1 || len(in.RuleUIDs) != 2 {
		t.Fatalf("uids not normalized: %+v %+v %+v", in.LevelUIDs, in.DatasourceUIDs, in.RuleUIDs)
	}
}

func TestNormalizeWindowRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		in   bo.AlertMaintenanceWindow
	}{
		{name: "empty name", in: func() bo.AlertMaintenanceWindow { w := validWindow(); w.Name = " "; return w }()},
		{name: "invalid labels", in: func() bo.AlertMaintenanceWindow { w := validWindow(); w.Labels = "[]"; return w }()},
		{name: "invalid label value", in: func() bo.AlertMaintenanceWindow { w := validWindow(); w.Labels = `{"env":1}`; return w }()},
		{name: "missing start", in: func() bo.AlertMaintenanceWindow { w := validWindow(); w.StartsAt = time.Time{}; return w }()},
		{name: "end before start", in: func() bo.AlertMaintenanceWindow {
			w := validWindow()
			w.EndsAt = w.StartsAt.Add(-time.Minute)
			return w
		}()},
		{name: "end equals start", in: func() bo.AlertMaintenanceWindow { w := validWindow(); w.EndsAt = w.StartsAt; return w }()},
	}
	for _, item := range cases {
		if err := normalizeWindow(&item.in); err == nil {
			t.Fatalf("%s: expected error", item.name)
		}
	}
}

func TestNormalizeWindowEmptyLabels(t *testing.T) {
	in := validWindow()
	in.Labels = "  "
	if err := normalizeWindow(&in); err != nil {
		t.Fatalf("empty labels rejected: %v", err)
	}
	if in.Labels != "{}" {
		t.Fatalf("labels = %q, want empty object", in.Labels)
	}
}
