// Package day provides a timer that matches the day of the month.
package day

import (
	"fmt"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*dayRange)(nil)

type dayRange struct {
	start, end int
}

// NewDayRange creates a timer that matches the day of the month
//
//	Example: NewDayRange(1, 31) means in 1, 2, ..., 31
//	Example: NewDayRange(1, 3) means in 1, 2, 3
//	Example: NewDayRange(31, 10) means in 31, 1, 2, ..., 10
func NewDayRange(start, end int) (timer.Timer, error) {
	if start > 31 || start < 1 || end > 31 || end < 1 {
		return nil, fmt.Errorf("start must be between 1 and 31, end must be between 1 and 31, but got start: %d, end: %d", start, end)
	}
	return &dayRange{start: start, end: end}, nil
}

func (d *dayRange) Match(now time.Time) bool {
	day := now.Day()
	if day < 1 || day > 31 {
		return false
	}
	if d.start <= d.end {
		return day >= d.start && day <= d.end
	}
	return day >= d.start || day <= d.end
}
