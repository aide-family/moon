package week

import (
	"slices"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*weekIn)(nil)

type weekIn struct {
	weeks []time.Weekday
}

// NewWeekIn creates a timer that matches the week of the year
//
//	Example: NewWeekIn(time.Monday, time.Wednesday, time.Friday) means in Monday, Wednesday or Friday
//	Example: NewWeekIn(time.Sunday, time.Monday) means in Sunday or Monday
func NewWeekIn(weeks ...time.Weekday) (timer.Timer, error) {
	weeks = slices.Compact(weeks)
	slices.Sort(weeks)
	return &weekIn{weeks: weeks}, nil
}

func (w *weekIn) Match(now time.Time) bool {
	return slices.Contains(w.weeks, now.Weekday())
}
