package timex

import (
	"strings"
	"sync"
	"time"
)

var location = "Asia/Shanghai"
var setLocationOnce sync.Once
var local, _ = time.LoadLocation(location)

func SetLocation(loc string) {
	setLocationOnce.Do(func() {
		location = loc
		var err error
		local, err = time.LoadLocation(loc)
		if err != nil {
			panic(err)
		}
	})
}

func GetLocation() *time.Location {
	if local == nil {
		panic("location is not set")
	}
	return local
}

func Now() time.Time {
	return time.Now().In(GetLocation())
}

func Format(t time.Time) string {
	return t.Format(time.DateTime)
}

func Parse(t string) (time.Time, error) {
	if strings.TrimSpace(t) == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation(time.DateTime, t, GetLocation())
}

func ParseX(t string) time.Time {
	parsed, err := Parse(t)
	if err != nil {
		return time.Time{}
	}
	return parsed
}
