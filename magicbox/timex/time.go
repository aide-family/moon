// Package timex provides time-related utilities.
package timex

import (
	"time"

	"github.com/bwmarrin/snowflake"
)

// StartOfDay returns the start of the day for the given time.
func StartOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}

// StartOfWeek returns the start of the week for the given time.
func StartOfWeek(t time.Time) time.Time {
	offset := int(time.Monday - t.Weekday())
	if offset > 0 {
		offset -= 7
	}
	return t.AddDate(0, 0, offset)
}

// FormatTime returns the formatted time for the given time.
func FormatTime(t *time.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Format(time.DateTime)
}

func TimeFromID(id snowflake.ID) time.Time {
	return time.UnixMilli(id.Time())
}
