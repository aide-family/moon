package oncallrotation

import (
	"testing"
	"time"

	"github.com/aide-family/moon/app/alioth/pkg/bo"
	"github.com/aide-family/moon/app/alioth/pkg/oncall"
)

// 个人视图的班次来自同一次排班计算，只是按成员过滤：
// 每天轮换、两人轮值，我应当拿到每隔一天的那几班，并按开始时间升序。
func TestMyShiftsOfFiltersByMember(t *testing.T) {
	rotation := bo.OnCallRotation{
		UID:         "rot-1",
		Name:        "后端值班",
		Timezone:    "UTC",
		Period:      oncall.RotationPeriodDaily,
		StartsAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		HandoffHour: 0,
		Members: []bo.OnCallRotationMember{
			{MemberUID: "m1"},
			{MemberUID: "m2"},
		},
	}
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	items, err := myShiftsOf(rotation, "m1", now, 4)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) < 2 {
		t.Fatalf("expected at least two shifts for m1, got %d", len(items))
	}
	for i, item := range items {
		if item.RotationUID != "rot-1" || item.RotationName != "后端值班" || item.Timezone != "UTC" {
			t.Fatalf("shift does not carry rotation info: %+v", item)
		}
		if !item.EndAt.After(item.StartAt) {
			t.Fatalf("shift end must be after start: %+v", item)
		}
		if i > 0 && !item.StartAt.After(items[i-1].StartAt) {
			t.Fatalf("shifts must be ascending by start time: %+v", items)
		}
	}
	if !items[0].Current {
		t.Fatalf("now falls inside m1's first shift, it should be current: %+v", items[0])
	}
	for _, item := range items[1:] {
		if item.Current {
			t.Fatalf("only the shift containing now should be current: %+v", item)
		}
	}

	// 同一时刻 m2 不在值班：过滤后没有当前班次，也不该混进 m1 的班次。
	other, err := myShiftsOf(rotation, "m2", now, 4)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range other {
		if item.Current {
			t.Fatalf("m2 should not be current at this instant: %+v", item)
		}
	}
}

// 不在轮换成员列表里的人拿到空列表：个人视图显示空态，而不是报错。
func TestMyShiftsOfUnknownMember(t *testing.T) {
	rotation := bo.OnCallRotation{
		UID:         "rot-1",
		Name:        "后端值班",
		Timezone:    "Asia/Shanghai",
		Period:      oncall.RotationPeriodDaily,
		StartsAt:    time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		HandoffHour: 9,
		Members:     []bo.OnCallRotationMember{{MemberUID: "m1"}},
	}
	items, err := myShiftsOf(rotation, "outsider", time.Now().UTC(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 0 {
		t.Fatalf("outsider should have no shifts, got %d", len(items))
	}
}
