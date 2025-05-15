package timer

import (
	"slices"
	"time"

	"github.com/moon-monitor/moon/pkg/merr"
)

// Ensure that the daysOfWeek type implements the Matcher interface.
var _ Matcher = (*daysOfWeek)(nil)

// NewDaysOfWeek creates and returns a new instance of daysOfWeek, which implements the Matcher interface.
// rule: A slice of integers representing valid days of the week (Sunday=0 to Saturday=6).
// Returns:
//   - A Matcher interface instance that can be used to match specific days of the week.
//   - An error if any day in the rule is invalid (not between Sunday and Saturday).
//
// This function validates each day in the rule to ensure it falls within the valid range of weekdays.
// If an invalid day is found, an error is returned.
func NewDaysOfWeek(rule []int) (Matcher, error) {
	days := make([]time.Weekday, 0, len(rule))
	for _, day := range rule {
		if day < int(time.Sunday) || day > int(time.Saturday) {
			return nil, merr.ErrorParamsError("invalid days of week: %v", rule)
		}
		days = append(days, time.Weekday(day))
	}
	return &daysOfWeek{Days: days}, nil
}

// daysOfWeek is a struct that holds a slice of valid weekdays.
// It implements the Match method to check if a given time matches any of the specified weekdays.
type daysOfWeek struct {
	Days []time.Weekday `json:"days"`
}

// Match checks if the given time 't' matches any of the weekdays stored in the daysOfWeek instance.
// t: A time.Time instance representing the time to check.
// Returns true if the weekday of 't' is contained in the Days slice; otherwise, returns false.
func (d *daysOfWeek) Match(t time.Time) bool {
	return slices.Contains(d.Days, t.Weekday())
}
