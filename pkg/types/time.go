package types

import (
	"time"
)

type Time time.Time

// String implements Stringer interface
func (t *Time) String() string {
	if t == nil {
		return "-"
	}
	return time.Time(*t).Format(time.DateTime)
}

// Unix implements Unix interface
func (t *Time) Unix() int64 {
	if t == nil {
		return 0
	}
	return time.Time(*t).Unix()
}
