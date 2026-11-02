// Package oncall 提供 OnCall 接收时段与值班轮换的纯计算。
// 计算不访问数据库：调用方把轮换配置与成员顺序传进来，结果可单测。
package oncall

import (
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/merr"
)

// 值班轮换周期。
const (
	RotationPeriodDaily  = "DAILY"
	RotationPeriodWeekly = "WEEKLY"
)

// RotationMember 是参与轮转的成员标识。展示信息由调用方按 MemberUID 补齐。
type RotationMember struct {
	MemberUID string
}

// Rotation 是一次排班计算需要的轮换配置。Members 的顺序即轮转顺序。
type Rotation struct {
	// Timezone 是 IANA 时区名，handoff_hour 按该时区解释。
	Timezone string
	// Period 取 DAILY 或 WEEKLY。
	Period string
	// StartsAt 是轮换起点，第一班从该时刻开始。
	StartsAt time.Time
	// HandoffHour 是本地时区的交班时刻，0–23；交班前的时刻属于上一班。
	HandoffHour int
	Members     []RotationMember
}

// Shift 是一个班次。StartAt/EndAt 是绝对时间，按轮换时区交班。
type Shift struct {
	StartAt time.Time
	EndAt   time.Time
	Member  RotationMember
	// Index 是班次序号，从轮换起点所在班次开始计 0。
	Index int
	// Current 表示该班次包含查询时刻（仅 Shifts 返回时设置）。
	Current bool
}

// periodDays 返回一个周期包含的自然日数。
func (r Rotation) periodDays() (int, error) {
	switch r.Period {
	case RotationPeriodDaily:
		return 1, nil
	case RotationPeriodWeekly:
		return 7, nil
	default:
		return 0, merr.Invalid("invalid rotation period")
	}
}

// location 返回轮换时区；非法时区返回客户端可读错误。
func (r Rotation) location() (*time.Location, error) {
	if strings.TrimSpace(r.Timezone) == "" {
		return nil, merr.Invalid("invalid rotation timezone")
	}
	return LoadLocation(r.Timezone)
}

// shiftStart 返回包含 t 的班次开始时刻：本地时区 handoff_hour 交班，
// WEEKLY 时交班日固定为 starts_at 的星期几；交班前的时刻仍属于上一班。
func (r Rotation) shiftStart(t time.Time, loc *time.Location, periodDays int) time.Time {
	local := t.In(loc)
	boundary := time.Date(local.Year(), local.Month(), local.Day(), r.HandoffHour, 0, 0, 0, loc)
	if local.Before(boundary) {
		boundary = boundary.AddDate(0, 0, -1)
	}
	if periodDays > 1 {
		anchorWeekday := r.StartsAt.In(loc).Weekday()
		back := (int(boundary.Weekday()) - int(anchorWeekday) + 7) % 7
		boundary = boundary.AddDate(0, 0, -back)
		if local.Before(boundary) {
			boundary = boundary.AddDate(0, 0, -periodDays)
		}
	}
	return boundary
}

// dateOrdinal 把本地日历日转成整数序号，跨夏令时比较仍按自然日计算。
func dateOrdinal(value time.Time) int64 {
	year, month, day := value.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC).Unix() / 86400
}

// shiftIndex 返回包含 t 的班次序号：从 starts_at 所在班次开始，按周期计整周期数。
// t 早于 starts_at 时返回 false。
func (r Rotation) shiftIndex(t time.Time, loc *time.Location, periodDays int) (int, bool) {
	if t.Before(r.StartsAt) {
		return 0, false
	}
	anchor := r.shiftStart(r.StartsAt, loc, periodDays)
	current := r.shiftStart(t, loc, periodDays)
	days := dateOrdinal(current.In(loc)) - dateOrdinal(anchor.In(loc))
	if days < 0 {
		days = 0
	}
	return int(days / int64(periodDays)), true
}

// valid 检查轮换配置是否足够排班。
func (r Rotation) valid() (*time.Location, int, error) {
	if len(r.Members) == 0 {
		return nil, 0, merr.Invalid("rotation has no members")
	}
	if r.HandoffHour < 0 || r.HandoffHour > 23 {
		return nil, 0, merr.Invalid("invalid rotation handoff hour")
	}
	loc, err := r.location()
	if err != nil {
		return nil, 0, err
	}
	periodDays, err := r.periodDays()
	if err != nil {
		return nil, 0, err
	}
	return loc, periodDays, nil
}

// MemberAt 返回 at 时刻的当班成员；轮换尚未开始时返回 false。
func (r Rotation) MemberAt(at time.Time) (RotationMember, bool) {
	loc, periodDays, err := r.valid()
	if err != nil {
		return RotationMember{}, false
	}
	index, ok := r.shiftIndex(at, loc, periodDays)
	if !ok {
		return RotationMember{}, false
	}
	return r.Members[index%len(r.Members)], true
}

// CurrentShift 返回包含 at 时刻的班次。
func (r Rotation) CurrentShift(at time.Time) (Shift, bool) {
	loc, periodDays, err := r.valid()
	if err != nil {
		return Shift{}, false
	}
	index, ok := r.shiftIndex(at, loc, periodDays)
	if !ok {
		return Shift{}, false
	}
	start := r.shiftStart(at, loc, periodDays)
	return Shift{
		StartAt: start, EndAt: start.AddDate(0, 0, periodDays),
		Member: r.Members[index%len(r.Members)], Index: index,
	}, true
}

// Shifts 返回 [from, to) 覆盖的班次，包含 from 所在的当前班次，最多 limit 个。
// limit <= 0 时按 64 处理，避免异常参数导致无限循环。
func (r Rotation) Shifts(from, to time.Time, limit int) ([]Shift, error) {
	if limit <= 0 {
		limit = 64
	}
	loc, periodDays, err := r.valid()
	if err != nil {
		return nil, err
	}
	if to.Before(from) {
		return nil, merr.Invalid("invalid rotation schedule range")
	}
	anchor := r.shiftStart(r.StartsAt, loc, periodDays)
	start := r.shiftStart(from, loc, periodDays)
	if start.Before(anchor) {
		start = anchor
	}
	shifts := make([]Shift, 0, limit)
	for start.Before(to) && len(shifts) < limit {
		// 班次序号直接按与起点的自然日差计算：首班可能早于 starts_at 的交班边界，
		// 不能用「早于起点就无班次」的 shiftIndex 判断。
		index := int(dateOrdinal(start.In(loc))-dateOrdinal(anchor.In(loc))) / periodDays
		// 第一班的开始时间不早于轮换起点：起点落在交班周期中间时，
		// 只展示从起点开始的那一段，结束时间仍是下一个交班时刻。
		displayStart := start
		if displayStart.Before(r.StartsAt) {
			displayStart = r.StartsAt
		}
		shift := Shift{
			StartAt: displayStart,
			EndAt:   start.AddDate(0, 0, periodDays),
			Member:  r.Members[index%len(r.Members)],
			Index:   index,
		}
		if !from.Before(shift.StartAt) && from.Before(shift.EndAt) {
			shift.Current = true
		}
		shifts = append(shifts, shift)
		start = shift.EndAt
	}
	return shifts, nil
}
