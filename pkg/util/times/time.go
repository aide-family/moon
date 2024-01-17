package times

import (
	"time"
)

type UnixTime interface {
	Unix() int64
}

const ParseLayout = "2006-01-02T15:04:05Z07:00"

func ParseAlertTime(timeStr string) time.Time {
	t, _ := time.Parse(ParseLayout, timeStr)
	return t
}

func ParseAlertTimeUnix(timeStr string) int64 {
	t, err := time.Parse(ParseLayout, timeStr)
	if err != nil {
		return 0
	}
	return t.Unix()
}

// TimeToUnix convert time.Time to unix timestamp
func TimeToUnix(t UnixTime) int64 {
	if t == nil {
		return 0
	}

	return t.Unix()
}
