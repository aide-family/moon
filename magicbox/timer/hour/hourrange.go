// Package hour provides a timer that matches the hour of the day.
package hour

import (
	"fmt"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*hourRange)(nil)

// NewHourRange creates an hourly range timer where start and end are integers from 0-23, and start must be less than or equal to end
//
//	Example: NewHourRange(10, 12) means between 10:00 and 12:00
//	Example: NewHourRange(10, 1) means between 10:00 and 1:00
func NewHourRange(start, end int) (timer.Timer, error) {
	if start > 23 || start < 0 || end > 23 || end < 0 {
		return nil, fmt.Errorf("start must be between 0 and 23, end must be between 0 and 23, but got start: %d, end: %d", start, end)
	}
	return &hourRange{
		start: start,
		ent:   end,
	}, nil
}

type hourRange struct {
	start, ent int
}

func (h *hourRange) Match(time time.Time) bool {
	nowHour := time.Hour()
	if h.start <= h.ent {
		return nowHour >= h.start && nowHour <= h.ent
	}
	return nowHour >= h.start || nowHour <= h.ent
}
