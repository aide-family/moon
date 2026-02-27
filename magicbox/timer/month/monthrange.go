// Package month provides a timer that matches the month of the year.
package month

import (
	"fmt"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*monthRange)(nil)

type monthRange struct {
	start, end time.Month
}

// NewMonthRange creates a timer that matches the month of the year
//
//	Example: NewMonthRange(time.January, time.March) means between January and March
//	Example: NewMonthRange(time.November, time.January) means between November and January
func NewMonthRange(start, end time.Month) (timer.Timer, error) {
	if start > time.December || start < time.January || end > time.December || end < time.January {
		return nil, fmt.Errorf("start must be between January and December, end must be between January and December, but got start: %s, end: %s", start, end)
	}
	return &monthRange{
		start: start,
		end:   end,
	}, nil
}

func (m *monthRange) Match(now time.Time) bool {
	nowMonth := now.Month()
	if nowMonth < time.January || nowMonth > time.December {
		return false
	}
	if m.start <= m.end {
		return nowMonth >= m.start && nowMonth <= m.end
	}
	return nowMonth >= m.start || nowMonth <= m.end
}
