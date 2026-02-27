package month

import (
	"fmt"
	"slices"
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*monthIn)(nil)

type monthIn struct {
	months []time.Month
}

// NewMonthIn creates a timer that matches the month of the year
//
//	Example: NewMonthIn(time.January, time.March, time.May) means in January, March or May
//	Example: NewMonthIn(time.November, time.January) means in November or January
func NewMonthIn(months ...time.Month) (timer.Timer, error) {
	months = slices.Compact(months)
	slices.Sort(months)

	for _, month := range months {
		if month < time.January || month > time.December {
			return nil, fmt.Errorf("month must be between January and December, but got %s", month)
		}
	}
	return &monthIn{
		months: months,
	}, nil
}

func (m *monthIn) Match(now time.Time) bool {
	nowMonth := now.Month()
	return slices.Contains(m.months, nowMonth)
}
