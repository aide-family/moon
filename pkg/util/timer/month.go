package timer

import (
	"time"

	"github.com/moon-monitor/moon/pkg/merr"
)

// Matcher is an interface for objects that can match a given time.Time against a rule.
var _ Matcher = (*dayOfMonths)(nil)
var _ Matcher = (*month)(nil)

// NewDayOfMonths creates a Matcher that matches if the day of the month
// falls within the specified start and end days, inclusive.
// It returns an error if the rule is invalid.
// The rule is represented as a slice of two integers,
// indicating the start and end days of the month, respectively.
func NewDayOfMonths(rule []int) (Matcher, error) {
	if len(rule) != 2 {
		return nil, merr.ErrorParamsError("invalid day of months: %v", rule)
	}
	start := rule[0]
	end := rule[1]
	if start < 1 || start > 31 || end < 1 || end > 31 {
		return nil, merr.ErrorParamsError("invalid day of months: %v", rule)
	}
	return &dayOfMonths{Start: start, End: end}, nil
}

// dayOfMonths is a struct representing a matching rule for days of the month.
type dayOfMonths struct {
	Start int `json:"start"` // Valid values: 1-31
	End   int `json:"end"`   // Valid values: 1-31
}

// Match checks if the given time.Time matches the day of month rule.
// If Start is greater than End, it checks if the day is at the start or end of the month.
func (m *dayOfMonths) Match(t time.Time) bool {
	d := t.Day()
	if m.Start > m.End {
		return d >= m.Start || d <= m.End
	}
	return d >= m.Start && d <= m.End
}

// NewMonth creates a Matcher that matches if the month
// falls within the specified start and end months, inclusive.
// It returns an error if the rule is invalid.
// The rule is represented as a slice of two integers,
// indicating the start and end months, respectively.
func NewMonth(rule []int) (Matcher, error) {
	if len(rule) != 2 {
		return nil, merr.ErrorParamsError("invalid month: %v", rule)
	}
	start := rule[0]
	end := rule[1]
	if start < 1 || start > 12 || end < 1 || end > 12 {
		return nil, merr.ErrorParamsError("invalid month: %v", rule)
	}
	return &month{Start: start, End: end}, nil
}

// month is a struct representing a matching rule for months.
type month struct {
	Start int `json:"start"` // Valid values: 1-12
	End   int `json:"end"`   // Valid values: 1-12
}

// Match checks if the given time.Time matches the month rule.
// If Start is greater than End, it checks if the month is at the start or end of the year.
func (m *month) Match(t time.Time) bool {
	mon := t.Month()
	if m.Start > m.End {
		return mon >= time.Month(m.Start) || mon <= time.Month(m.End)
	}
	return mon >= time.Month(m.Start) && mon <= time.Month(m.End)
}
