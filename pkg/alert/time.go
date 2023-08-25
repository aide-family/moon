package alert

import "time"

const ParseLayout = "2006-01-02T15:04:05Z07:00"

func ParseTime(timeStr string) time.Time {
	t, _ := time.Parse(ParseLayout, timeStr)
	return t
}
