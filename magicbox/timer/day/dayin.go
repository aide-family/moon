package day

import (
	"fmt"
	"slices"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*dayIn)(nil)

type dayIn struct {
	days []int
}

// NewDayIn creates a timer that matches the day of the month
//
//	Example: NewDayIn(1, 3, 5) means in 1, 3, 5
//	Example: NewDayIn(1, 3) means in 1, 3
//	Example: NewDayIn(31, 10) means in 31, 1, 2, ..., 10
func NewDayIn(days ...int) (timer.Timer, error) {
	days = slices.Compact(days)
	slices.Sort(days)

	for _, day := range days {
		if day < 1 || day > 31 {
			return nil, fmt.Errorf("day must be between 1 and 31, but got %d", day)
		}
	}
	return &dayIn{days: days}, nil
}

func (d *dayIn) Match(now time.Time) bool {
	day := now.Day()
	if day < 1 || day > 31 {
		return false
	}
	return slices.Contains(d.days, day)
}
