// Package week provides a timer that matches the week of the year.
package week

import (
	"fmt"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*weekRange)(nil)

type weekRange struct {
	start, end time.Weekday
}

// NewWeekRange creates a timer that matches the week of the year
//
//	Example: NewWeekRange(time.Monday, time.Wednesday) means between Monday and Wednesday
//	Example: NewWeekRange(time.Sunday, time.Monday) means between Sunday and Monday
func NewWeekRange(start, end time.Weekday) (timer.Timer, error) {
	if start > time.Saturday || start < time.Sunday || end < time.Sunday || end > time.Saturday {
		return nil, fmt.Errorf("start must be between Monday and Sunday, end must be between Monday and Sunday, but got start: %s, end: %s", start, end)
	}
	return &weekRange{start: start, end: end}, nil
}

func (w *weekRange) Match(now time.Time) bool {
	week := now.Weekday()
	if w.start <= w.end {
		return week >= w.start && week <= w.end
	}
	return week >= w.start || week <= w.end
}
