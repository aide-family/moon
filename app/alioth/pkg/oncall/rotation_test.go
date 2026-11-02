package oncall

import (
	"testing"
	"time"
)

func testLocation(t *testing.T, name string) *time.Location {
	t.Helper()
	loc, err := time.LoadLocation(name)
	if err != nil {
		t.Skipf("timezone %s unavailable: %v", name, err)
	}
	return loc
}

func testRotation(t *testing.T, timezone, period string, starts time.Time, handoff int, members ...string) Rotation {
	t.Helper()
	items := make([]RotationMember, 0, len(members))
	for _, member := range members {
		items = append(items, RotationMember{MemberUID: member})
	}
	return Rotation{
		Timezone: timezone, Period: period, StartsAt: starts,
		HandoffHour: handoff, Members: items,
	}
}

func requireMember(t *testing.T, rotation Rotation, at time.Time, want string) {
	t.Helper()
	member, ok := rotation.MemberAt(at)
	if !ok {
		t.Fatalf("MemberAt(%s) = no member", at)
	}
	if member.MemberUID != want {
		t.Fatalf("MemberAt(%s) = %s, want %s", at, member.MemberUID, want)
	}
}

// TestRotationDailyHandoff 验证每日轮换：交班前的时刻属于上一班，交班后切到下一位成员。
func TestRotationDailyHandoff(t *testing.T) {
	loc := testLocation(t, "Asia/Shanghai")
	starts := time.Date(2026, 1, 1, 9, 0, 0, 0, loc)
	rotation := testRotation(t, "Asia/Shanghai", RotationPeriodDaily, starts, 9, "m1", "m2", "m3")

	requireMember(t, rotation, starts, "m1")
	requireMember(t, rotation, time.Date(2026, 1, 2, 8, 59, 0, 0, loc), "m1")
	requireMember(t, rotation, time.Date(2026, 1, 2, 9, 0, 0, 0, loc), "m2")
	requireMember(t, rotation, time.Date(2026, 1, 2, 23, 0, 0, 0, loc), "m2")
	requireMember(t, rotation, time.Date(2026, 1, 3, 9, 30, 0, 0, loc), "m3")
	requireMember(t, rotation, time.Date(2026, 1, 4, 9, 0, 0, 0, loc), "m1")
}

// TestRotationHandoffBeforeStartsAt 起点晚于当天交班时刻时，第一班仍从起点开始；
// 第二天的交班时刻才切人。
func TestRotationHandoffBeforeStartsAt(t *testing.T) {
	loc := testLocation(t, "Asia/Shanghai")
	starts := time.Date(2026, 1, 1, 10, 0, 0, 0, loc)
	rotation := testRotation(t, "Asia/Shanghai", RotationPeriodDaily, starts, 8, "m1", "m2")

	requireMember(t, rotation, starts, "m1")
	requireMember(t, rotation, time.Date(2026, 1, 2, 7, 59, 0, 0, loc), "m1")
	requireMember(t, rotation, time.Date(2026, 1, 2, 8, 0, 0, 0, loc), "m2")
}

// TestRotationWeekly 验证每周轮换：交班日固定为起点星期几，交班前仍是上一周的班。
func TestRotationWeekly(t *testing.T) {
	loc := testLocation(t, "Asia/Shanghai")
	// 2026-01-05 是星期一。
	starts := time.Date(2026, 1, 5, 9, 0, 0, 0, loc)
	rotation := testRotation(t, "Asia/Shanghai", RotationPeriodWeekly, starts, 9, "m1", "m2", "m3")

	requireMember(t, rotation, starts, "m1")
	requireMember(t, rotation, time.Date(2026, 1, 12, 8, 0, 0, 0, loc), "m1")
	requireMember(t, rotation, time.Date(2026, 1, 12, 9, 0, 0, 0, loc), "m2")
	requireMember(t, rotation, time.Date(2026, 1, 19, 9, 0, 0, 0, loc), "m3")
	requireMember(t, rotation, time.Date(2026, 1, 26, 9, 0, 0, 0, loc), "m1")
}

// TestRotationAcrossDST 验证跨夏令时仍按自然日推进：3 月 8 日的班次只有 23 小时，
// 但 3 月 9 日交班后仍然切换到下一位成员。
func TestRotationAcrossDST(t *testing.T) {
	loc := testLocation(t, "America/New_York")
	starts := time.Date(2026, 3, 7, 0, 0, 0, 0, loc)
	rotation := testRotation(t, "America/New_York", RotationPeriodDaily, starts, 0, "m1", "m2", "m3")

	requireMember(t, rotation, time.Date(2026, 3, 7, 12, 0, 0, 0, loc), "m1")
	requireMember(t, rotation, time.Date(2026, 3, 8, 12, 0, 0, 0, loc), "m2")
	requireMember(t, rotation, time.Date(2026, 3, 9, 0, 0, 0, 0, loc), "m3")
	shifts, err := rotation.Shifts(time.Date(2026, 3, 7, 0, 0, 0, 0, loc), time.Date(2026, 3, 10, 0, 0, 0, 0, loc), 0)
	if err != nil {
		t.Fatalf("shifts: %v", err)
	}
	if len(shifts) != 3 {
		t.Fatalf("shifts = %d, want 3", len(shifts))
	}
	for _, shift := range shifts {
		if hour := shift.StartAt.In(loc).Hour(); hour != 0 {
			t.Fatalf("shift start local hour = %d, want 0", hour)
		}
	}
}

// TestRotationBeforeStart 轮换开始前没有当班成员。
func TestRotationBeforeStart(t *testing.T) {
	loc := testLocation(t, "Asia/Shanghai")
	starts := time.Date(2026, 1, 1, 9, 0, 0, 0, loc)
	rotation := testRotation(t, "Asia/Shanghai", RotationPeriodDaily, starts, 9, "m1")
	if _, ok := rotation.MemberAt(starts.Add(-time.Minute)); ok {
		t.Fatal("rotation before starts_at should have no member")
	}
}

// TestRotationShifts 验证未来班次：包含当前班次并标记 Current，数量受 limit 限制。
func TestRotationShifts(t *testing.T) {
	loc := testLocation(t, "Asia/Shanghai")
	starts := time.Date(2026, 1, 1, 9, 0, 0, 0, loc)
	rotation := testRotation(t, "Asia/Shanghai", RotationPeriodDaily, starts, 9, "m1", "m2")
	now := time.Date(2026, 1, 2, 12, 0, 0, 0, loc)
	shifts, err := rotation.Shifts(now, now.AddDate(0, 0, 3), 0)
	if err != nil {
		t.Fatalf("shifts: %v", err)
	}
	if len(shifts) != 4 {
		t.Fatalf("shifts = %d, want 4", len(shifts))
	}
	if !shifts[0].Current || shifts[0].Member.MemberUID != "m2" {
		t.Fatalf("first shift = current %v member %s", shifts[0].Current, shifts[0].Member.MemberUID)
	}
	if shifts[1].Current {
		t.Fatal("only the shift containing now should be current")
	}
	if got := shifts[0].EndAt.In(loc).Hour(); got != 9 {
		t.Fatalf("shift end local hour = %d, want 9", got)
	}
}

// TestRotationFutureStart 起点在未来时，排班从起点开始展示，首班结束于下一个交班时刻。
func TestRotationFutureStart(t *testing.T) {
	loc := testLocation(t, "Asia/Shanghai")
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, loc)
	starts := time.Date(2026, 1, 3, 10, 0, 0, 0, loc)
	rotation := testRotation(t, "Asia/Shanghai", RotationPeriodDaily, starts, 8, "m1", "m2")
	shifts, err := rotation.Shifts(now, now.AddDate(0, 0, 5), 0)
	if err != nil {
		t.Fatalf("shifts: %v", err)
	}
	if len(shifts) == 0 {
		t.Fatal("future rotation should return shifts")
	}
	if !shifts[0].StartAt.Equal(starts) {
		t.Fatalf("first shift start = %s, want starts_at %s", shifts[0].StartAt, starts)
	}
	if hour := shifts[0].EndAt.In(loc).Hour(); hour != 8 {
		t.Fatalf("first shift end local hour = %d, want 8", hour)
	}
	if shifts[0].Member.MemberUID != "m1" {
		t.Fatalf("first shift member = %s, want m1", shifts[0].Member.MemberUID)
	}
	if shifts[0].Current {
		t.Fatal("future shift must not be marked current")
	}
}

// TestRotationInvalidConfig 非法周期、时区、空成员、越界交班时刻都要拒绝。
func TestRotationInvalidConfig(t *testing.T) {
	loc := testLocation(t, "Asia/Shanghai")
	starts := time.Date(2026, 1, 1, 9, 0, 0, 0, loc)
	cases := []Rotation{
		{Timezone: "Asia/Shanghai", Period: "MONTHLY", StartsAt: starts, Members: []RotationMember{{MemberUID: "m1"}}},
		{Timezone: "Not/AZone", Period: RotationPeriodDaily, StartsAt: starts, Members: []RotationMember{{MemberUID: "m1"}}},
		{Timezone: "Asia/Shanghai", Period: RotationPeriodDaily, StartsAt: starts},
		{Timezone: "Asia/Shanghai", Period: RotationPeriodDaily, StartsAt: starts, HandoffHour: 24, Members: []RotationMember{{MemberUID: "m1"}}},
	}
	for index, rotation := range cases {
		if _, ok := rotation.MemberAt(starts); ok {
			t.Fatalf("case %d: invalid rotation returned a member", index)
		}
		if _, err := rotation.Shifts(starts, starts.AddDate(0, 0, 1), 1); err == nil {
			t.Fatalf("case %d: invalid rotation returned shifts", index)
		}
	}
}
