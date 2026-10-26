package notification

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aide-family/Alioth/app/alioth/pkg/bo"
)

func validOnCall() bo.NotificationOnCall {
	return bo.NotificationOnCall{
		Name: "夜间值班",
		Groups: []bo.NotificationOnCallGroup{
			{GroupUID: "group-1"},
		},
	}
}

func TestNormalizeOnCall(t *testing.T) {
	in := bo.NotificationOnCall{
		Name:                  "  夜间值班  ",
		Labels:                `{ "job" : ".+", "alertname" : "HighCPU" }`,
		LevelUIDs:             []string{"level-1", " level-1 ", "level-2", ""},
		DatasourceUIDs:        []string{"ds-1", "ds-1"},
		DatasourceLevelUIDs:   []string{"dsl-1", " dsl-1 ", "dsl-2", ""},
		RepeatIntervalSeconds: 3600,
		Timezone:              "Asia/Shanghai",
		Groups: []bo.NotificationOnCallGroup{
			{
				GroupUID: " group-1 ",
				Windows: []bo.NotificationOnCallWindow{
					{Weekdays: []string{"FRIDAY", "MONDAY", "MONDAY"}, Start: "9:00", End: "18:00"},
					{Weekdays: []string{"SATURDAY"}, Start: "22:00", End: "06:00"},
				},
			},
			{GroupUID: "group-2"},
		},
	}
	if err := normalizeOnCall(&in); err != nil {
		t.Fatalf("valid on-call rejected: %v", err)
	}
	if in.Name != "夜间值班" {
		t.Fatalf("name = %q, want trimmed", in.Name)
	}
	if in.Labels != `{"alertname":"HighCPU","job":".+"}` {
		t.Fatalf("labels = %q, want compact JSON", in.Labels)
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
	if in.RepeatIntervalSeconds != 3600 {
		t.Fatalf("repeat interval seconds = %d, want 3600", in.RepeatIntervalSeconds)
	}
	if in.Timezone != "Asia/Shanghai" {
		t.Fatalf("timezone = %q", in.Timezone)
	}
	if len(in.Groups) != 2 || in.Groups[0].GroupUID != "group-1" {
		t.Fatalf("groups = %+v", in.Groups)
	}
	windows := in.Groups[0].Windows
	if len(windows) != 2 {
		t.Fatalf("windows = %+v, want 2", windows)
	}
	if len(windows[0].Weekdays) != 2 || windows[0].Weekdays[0] != "MONDAY" || windows[0].Weekdays[1] != "FRIDAY" {
		t.Fatalf("weekdays = %v, want sorted and deduplicated", windows[0].Weekdays)
	}
	if windows[0].Start != "09:00" || windows[0].End != "18:00" {
		t.Fatalf("window time = %s-%s, want padded HH:mm", windows[0].Start, windows[0].End)
	}
	if windows[1].Start != "22:00" || windows[1].End != "06:00" {
		t.Fatalf("cross-midnight window = %s-%s", windows[1].Start, windows[1].End)
	}
	if len(in.Groups[1].Windows) != 0 {
		t.Fatalf("group without windows = %+v, want empty", in.Groups[1].Windows)
	}
}

func TestNormalizeOnCallDefaultsTimezone(t *testing.T) {
	in := validOnCall()
	if err := normalizeOnCall(&in); err != nil {
		t.Fatalf("valid on-call rejected: %v", err)
	}
	if in.Timezone != "Asia/Shanghai" {
		t.Fatalf("timezone = %q, want default Asia/Shanghai", in.Timezone)
	}
	if in.Labels != "{}" {
		t.Fatalf("empty labels = %q, want {}", in.Labels)
	}
}

func TestNormalizeOnCallMergesDuplicateWindows(t *testing.T) {
	in := validOnCall()
	in.Groups[0].Windows = []bo.NotificationOnCallWindow{
		{Weekdays: []string{"MONDAY"}, Start: "09:00", End: "18:00"},
		{Weekdays: []string{"MONDAY"}, Start: "09:00", End: "18:00"},
	}
	if err := normalizeOnCall(&in); err != nil {
		t.Fatalf("valid on-call rejected: %v", err)
	}
	if len(in.Groups[0].Windows) != 1 {
		t.Fatalf("windows = %+v, want duplicate merged", in.Groups[0].Windows)
	}
}

func TestNormalizeOnCallAcceptsRepeatIntervalBounds(t *testing.T) {
	for _, seconds := range []int{0, maxRepeatIntervalSeconds} {
		in := validOnCall()
		in.RepeatIntervalSeconds = seconds
		if err := normalizeOnCall(&in); err != nil {
			t.Fatalf("repeat interval %d rejected: %v", seconds, err)
		}
		if in.RepeatIntervalSeconds != seconds {
			t.Fatalf("repeat interval seconds = %d, want %d", in.RepeatIntervalSeconds, seconds)
		}
	}
}

func TestNormalizeOnCallRejects(t *testing.T) {
	twentyOneWindows := make([]bo.NotificationOnCallWindow, 0, 21)
	for hour := 0; hour < 21; hour++ {
		twentyOneWindows = append(twentyOneWindows, bo.NotificationOnCallWindow{
			Weekdays: []string{"MONDAY"},
			Start:    fmt.Sprintf("%02d:00", hour),
			End:      fmt.Sprintf("%02d:30", hour),
		})
	}
	tooManyLevels := make([]string, 0, 201)
	for index := 0; index < 201; index++ {
		tooManyLevels = append(tooManyLevels, fmt.Sprintf("level-%d", index))
	}
	tooManyDatasourceLevels := make([]string, 0, 201)
	for index := 0; index < 201; index++ {
		tooManyDatasourceLevels = append(tooManyDatasourceLevels, fmt.Sprintf("dsl-%d", index))
	}
	tooManyGroups := make([]bo.NotificationOnCallGroup, 0, 51)
	for index := 0; index < 51; index++ {
		tooManyGroups = append(tooManyGroups, bo.NotificationOnCallGroup{GroupUID: fmt.Sprintf("group-%d", index)})
	}
	tooLongLabels := `{"job":"` + strings.Repeat("a", maxOnCallLabelsBytes) + `"}`

	cases := []struct {
		name   string
		mutate func(*bo.NotificationOnCall)
	}{
		{"missing name", func(in *bo.NotificationOnCall) { in.Name = "  " }},
		{"name too long", func(in *bo.NotificationOnCall) { in.Name = strings.Repeat("值", maxOnCallNameRunes+1) }},
		{"invalid labels", func(in *bo.NotificationOnCall) { in.Labels = `{"job":` }},
		{"invalid label matcher", func(in *bo.NotificationOnCall) { in.Labels = `{"job":"["}` }},
		{"labels too long", func(in *bo.NotificationOnCall) { in.Labels = tooLongLabels }},
		{"too many level uids", func(in *bo.NotificationOnCall) { in.LevelUIDs = tooManyLevels }},
		{"too many datasource level uids", func(in *bo.NotificationOnCall) { in.DatasourceLevelUIDs = tooManyDatasourceLevels }},
		{"negative repeat interval", func(in *bo.NotificationOnCall) { in.RepeatIntervalSeconds = -1 }},
		{"repeat interval above maximum", func(in *bo.NotificationOnCall) {
			in.RepeatIntervalSeconds = maxRepeatIntervalSeconds + 1
		}},
		{"no groups", func(in *bo.NotificationOnCall) { in.Groups = nil }},
		{"empty group uid", func(in *bo.NotificationOnCall) { in.Groups[0].GroupUID = " " }},
		{"duplicate group", func(in *bo.NotificationOnCall) {
			in.Groups = append(in.Groups, bo.NotificationOnCallGroup{GroupUID: "group-1"})
		}},
		{"too many groups", func(in *bo.NotificationOnCall) { in.Groups = tooManyGroups }},
		{"invalid timezone", func(in *bo.NotificationOnCall) { in.Timezone = "Mars/Olympus" }},
		{"window start equals end", func(in *bo.NotificationOnCall) {
			in.Groups[0].Windows = []bo.NotificationOnCallWindow{{Weekdays: []string{"MONDAY"}, Start: "09:00", End: "09:00"}}
		}},
		{"invalid weekday", func(in *bo.NotificationOnCall) {
			in.Groups[0].Windows = []bo.NotificationOnCallWindow{{Weekdays: []string{"FUNDAY"}, Start: "09:00", End: "18:00"}}
		}},
		{"window without weekday", func(in *bo.NotificationOnCall) {
			in.Groups[0].Windows = []bo.NotificationOnCallWindow{{Start: "09:00", End: "18:00"}}
		}},
		{"invalid window time", func(in *bo.NotificationOnCall) {
			in.Groups[0].Windows = []bo.NotificationOnCallWindow{{Weekdays: []string{"MONDAY"}, Start: "9:00", End: "25:00"}}
		}},
		{"too many windows", func(in *bo.NotificationOnCall) { in.Groups[0].Windows = twentyOneWindows }},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			in := validOnCall()
			testCase.mutate(&in)
			if err := normalizeOnCall(&in); err == nil {
				t.Fatalf("invalid on-call accepted: %+v", in)
			}
		})
	}
}
