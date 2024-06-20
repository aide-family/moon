package types

import (
	"database/sql/driver"
	"fmt"
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

func NewTime(t time.Time) *Time {
	return (*Time)(&t)
}

func NewTimeByString(s string, layout ...string) *Time {
	lay := time.DateTime
	if len(layout) > 0 {
		lay = layout[0]
	}
	t, err := time.ParseInLocation(lay, s, time.Local)
	if err != nil {
		return nil
	}
	return (*Time)(&t)
}

// Scan 现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (t *Time) Scan(value interface{}) error {
	switch value.(type) {
	case time.Time:
		*t = Time(value.(time.Time))
	case string:
		tt, err := time.ParseInLocation(time.DateTime, value.(string), time.Local)
		if err != nil {
			return err
		}
		*t = Time(tt)
	case nil:
		*t = Time(time.Time{})
	default:
		return fmt.Errorf("can not convert %v to Time", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口，Value
func (t Time) Value() (driver.Value, error) {
	return time.Time(t), nil
}
