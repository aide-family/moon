package timer

import (
	"slices"
	"time"
)

// Matcher is an interface for time matching.
var _ Matcher = (*hourRange)(nil)
var _ Matcher = (*hour)(nil)
var _ Matcher = (*hourMinuteRange)(nil)

// NewHourRange creates an hourRange Matcher based on the hour range rule.
// rule: A slice containing 2 integers representing the start and end hours.
// Returns a Matcher that matches the specified hour range, or an error if the rule is invalid.
func NewHourRange(rule []int) (Matcher, error) {
	if err := ValidateHourRange(rule); err != nil {
		return nil, err
	}
	return &hourRange{Start: rule[0], End: rule[1]}, nil
}

// hourRange is a struct for defining an hour range, implementing the Matcher interface.
type hourRange struct {
	Start int `json:"start"` // 0-23
	End   int `json:"end"`   // 0-23
}

// Match checks if the given time matches the hour range.
// t: The time to check.
// Returns true if t's hour is within the range, otherwise false.
func (h *hourRange) Match(t time.Time) bool {
	hh := t.Hour()
	if h.Start > h.End {
		return hh >= h.Start || hh <= h.End
	}
	return hh >= h.Start && hh <= h.End
}

// NewHour creates an hour Matcher based on the hour rule.
// rule: A slice containing the hours to match.
// Returns a Matcher that matches any of the specified hours.
func NewHour(rule []int) (Matcher, error) {
	if err := ValidateHour(rule); err != nil {
		return nil, err
	}
	return &hour{Hours: rule}, nil
}

// hour is a struct for defining specific hours, implementing the Matcher interface.
type hour struct {
	Hours []int `json:"hours"` // 0-23
}

// Match checks if the given time matches any of the specified hours.
// t: The time to check.
// Returns true if t's hour is in the Hours slice, otherwise false.
func (h *hour) Match(t time.Time) bool {
	hh := t.Hour()
	return slices.Contains(h.Hours, hh)
}

// NewHourMinuteRange creates an hourMinuteRange Matcher based on the start and end hour-minute rules.
// startHourMinute: The start hour and minute.
// endHourMinute: The end hour and minute.
// Returns a Matcher that matches the specified hour-minute range.
func NewHourMinuteRange(startHourMinute, endHourMinute HourMinute) (Matcher, error) {
	return &hourMinuteRange{Start: startHourMinute, End: endHourMinute}, nil
}

// NewHourMinuteRangeWithSlice creates an hourMinuteRange Matcher based on the start and end hour-minute rules.
// hour: A slice containing 4 integers representing the start hour, start minute, end hour, and end minute.
// Returns a Matcher that matches the specified hour-minute range.
func NewHourMinuteRangeWithSlice(hour []int) (Matcher, error) {
	if err := ValidateHourMinuteRange(hour); err != nil {
		return nil, err
	}
	startHour, startMinute := hour[0], hour[1]
	startHourMinute, err := NewHourMinute(startHour, startMinute)
	if err != nil {
		return nil, err
	}
	endHour, endMinute := hour[2], hour[3]
	endHourMinute, err := NewHourMinute(endHour, endMinute)
	if err != nil {
		return nil, err
	}
	return &hourMinuteRange{Start: *startHourMinute, End: *endHourMinute}, nil
}

// hourMinuteRange is a struct for defining an hour-minute range, implementing the Matcher interface.
type hourMinuteRange struct {
	Start HourMinute `json:"start"`
	End   HourMinute `json:"end"`
}

// Match checks if the given time matches the hour-minute range.
// t: The time to check.
// Returns true if t's hour and minute are within the range, otherwise false.
func (h *hourMinuteRange) Match(t time.Time) bool {
	hh := t.Hour()
	mm := t.Minute()

	checkHourRange := &HourMinute{Hour: hh, Minute: mm}
	if h.Start.GT(&h.End) {
		return h.End.LT(checkHourRange) || h.Start.GT(checkHourRange)
	}

	if h.Start.GT(checkHourRange) || h.End.LT(checkHourRange) {
		return false
	}
	return true
}

// NewHourMinute creates an HourMinute Matcher based on the hour and minute rule.
// hour: The hour to match.
// minute: The minute to match.
// Returns a Matcher that matches the specified hour and minute.
func NewHourMinute(hour, minute int) (*HourMinute, error) {
	if err := ValidateHourMinute(hour, minute); err != nil {
		return nil, err
	}
	return &HourMinute{Hour: hour, Minute: minute}, nil
}

// HourMinute is a struct for defining a specific hour and minute.
type HourMinute struct {
	Hour   int `json:"hour"`   // 0-23
	Minute int `json:"minute"` // 0-59
}

// GT checks if the current hourMinute is greater than the one provided.
// hourMinute: The hourMinute to compare with.
// Returns true if the current hourMinute is greater, otherwise false.
func (h *HourMinute) GT(hourMinute *HourMinute) bool {
	hh := h.Hour
	mm := h.Minute
	if hh > hourMinute.Hour {
		return true
	}
	if hh == hourMinute.Hour && mm > hourMinute.Minute {
		return true
	}
	return false
}

// LT checks if the current hourMinute is less than the one provided.
// hourMinute: The hourMinute to compare with.
// Returns true if the current hourMinute is less, otherwise false.
func (h *HourMinute) LT(hourMinute *HourMinute) bool {
	hh := h.Hour
	mm := h.Minute
	if hh < hourMinute.Hour {
		return true
	}
	if hh == hourMinute.Hour && mm < hourMinute.Minute {
		return true
	}
	return false
}
