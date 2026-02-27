package hour

import (
	"fmt"
	"slices"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*hourIn)(nil)

type hourIn struct {
	hours []int
}

// NewHourIn creates a timer that matches the hour of the day
//
//	Example: NewHourIn(10, 12) means in 10:00 or 12:00
//	Example: NewHourIn(10, 1) means in 10:00 or 1:00
func NewHourIn(hours ...int) (timer.Timer, error) {
	hours = slices.Compact(hours)
	slices.Sort(hours)

	for _, hour := range hours {
		if hour < 0 || hour > 23 {
			return nil, fmt.Errorf("hour must be between 0 and 23, but got %d", hour)
		}
	}
	return &hourIn{
		hours: hours,
	}, nil
}

func (h *hourIn) Match(time time.Time) bool {
	return slices.Contains(h.hours, time.Hour())
}
