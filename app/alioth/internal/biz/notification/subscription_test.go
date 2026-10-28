package notification

import (
	"fmt"
	"testing"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
)

func validSubscription() bo.NotificationSubscription {
	return bo.NotificationSubscription{
		Name:      "核心告警",
		GroupUIDs: []string{"group-1"},
	}
}

func TestNormalizeSubscriptionFields(t *testing.T) {
	in := bo.NotificationSubscription{
		Name:                  "  核心告警  ",
		LevelUIDs:             []string{"level-1", " level-1 ", "level-2", ""},
		DatasourceUIDs:        []string{"ds-1", "ds-1"},
		DatasourceLevelUIDs:   []string{"dsl-1", " dsl-1 ", "dsl-2", ""},
		GroupUIDs:             []string{"group-1", "group-1"},
		RepeatIntervalSeconds: 3600,
	}
	if err := normalizeSubscriptionFields(&in); err != nil {
		t.Fatalf("valid subscription rejected: %v", err)
	}
	if in.Name != "核心告警" {
		t.Fatalf("name = %q, want trimmed", in.Name)
	}
	if len(in.LevelUIDs) != 2 || in.LevelUIDs[0] != "level-1" || in.LevelUIDs[1] != "level-2" {
		t.Fatalf("level uids = %v, want deduplicated", in.LevelUIDs)
	}
	if len(in.DatasourceUIDs) != 1 || in.DatasourceUIDs[0] != "ds-1" {
		t.Fatalf("datasource uids = %v, want deduplicated", in.DatasourceUIDs)
	}
	if len(in.DatasourceLevelUIDs) != 2 || in.DatasourceLevelUIDs[0] != "dsl-1" || in.DatasourceLevelUIDs[1] != "dsl-2" {
		t.Fatalf("datasource level uids = %v, want deduplicated", in.DatasourceLevelUIDs)
	}
	if len(in.GroupUIDs) != 1 || in.GroupUIDs[0] != "group-1" {
		t.Fatalf("group uids = %v, want deduplicated", in.GroupUIDs)
	}
	if in.RepeatIntervalSeconds != 3600 {
		t.Fatalf("repeat interval seconds = %d, want 3600", in.RepeatIntervalSeconds)
	}
}

func TestNormalizeSubscriptionFieldsAcceptsRepeatIntervalBounds(t *testing.T) {
	for _, seconds := range []int{0, maxRepeatIntervalSeconds} {
		in := validSubscription()
		in.RepeatIntervalSeconds = seconds
		if err := normalizeSubscriptionFields(&in); err != nil {
			t.Fatalf("repeat interval %d rejected: %v", seconds, err)
		}
	}
}

func TestNormalizeSubscriptionFieldsRejects(t *testing.T) {
	tooManyLevels := make([]string, 0, 201)
	tooManyDatasources := make([]string, 0, 201)
	tooManyDatasourceLevels := make([]string, 0, 201)
	for index := 0; index < 201; index++ {
		tooManyLevels = append(tooManyLevels, fmt.Sprintf("level-%d", index))
		tooManyDatasources = append(tooManyDatasources, fmt.Sprintf("ds-%d", index))
		tooManyDatasourceLevels = append(tooManyDatasourceLevels, fmt.Sprintf("dsl-%d", index))
	}

	cases := []struct {
		name   string
		mutate func(*bo.NotificationSubscription)
	}{
		{"missing name", func(in *bo.NotificationSubscription) { in.Name = "  " }},
		{"too many level uids", func(in *bo.NotificationSubscription) { in.LevelUIDs = tooManyLevels }},
		{"too many datasource uids", func(in *bo.NotificationSubscription) { in.DatasourceUIDs = tooManyDatasources }},
		{"too many datasource level uids", func(in *bo.NotificationSubscription) { in.DatasourceLevelUIDs = tooManyDatasourceLevels }},
		{"negative repeat interval", func(in *bo.NotificationSubscription) { in.RepeatIntervalSeconds = -1 }},
		{"repeat interval above maximum", func(in *bo.NotificationSubscription) {
			in.RepeatIntervalSeconds = maxRepeatIntervalSeconds + 1
		}},
		{"no groups", func(in *bo.NotificationSubscription) { in.GroupUIDs = nil }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			in := validSubscription()
			testCase.mutate(&in)
			if err := normalizeSubscriptionFields(&in); err == nil {
				t.Fatalf("invalid subscription accepted: %+v", in)
			}
		})
	}
}
