package hour

import (
	"time"

	"github.com/aide-family/magicbox/timer"
)

var _ timer.Timer = (*hourMinuteRange)(nil)

type hourMinuteRange struct {
	start, end time.Time
}

// NewHourMinuteRange creates a timer that matches the hour and minute of the day
//
//	Example: NewHourMinuteRange(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)) means between 10:00 and 12:00
//	Example: NewHourMinuteRange(time.Date(2025, 1, 1, 10, 0, 0, 0, time.UTC), time.Date(2025, 1, 1, 1, 0, 0, 0, time.UTC)) means between 10:00 and 1:00
func NewHourMinuteRange(start, end time.Time) (timer.Timer, error) {
	return &hourMinuteRange{
		start: start,
		end:   end,
	}, nil
}

func (h *hourMinuteRange) Match(now time.Time) bool {
	year, month, day := now.Date()
	start := time.Date(year, month, day, h.start.Hour(), h.start.Minute(), 0, 0, now.Location())
	end := time.Date(year, month, day, h.end.Hour(), h.end.Minute(), 0, 0, now.Location())
	if now.Equal(start) || now.Equal(end) {
		return true
	}
	if start.After(end) {
		return now.After(start) || now.Before(end)
	}
	return now.After(start) && now.Before(end)
}
