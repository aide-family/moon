package types

import (
	"fmt"
	"time"

	"database/sql"
	"database/sql/driver"

	"google.golang.org/protobuf/types/known/durationpb"
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

var _ driver.Valuer = (*Duration)(nil)
var _ sql.Scanner = (*Duration)(nil)

type Duration struct {
	Duration *durationpb.Duration
}

func (d *Duration) Value() (driver.Value, error) {
	return int64(d.GetDuration().AsDuration()), nil
}

func (d *Duration) Scan(src any) error {
	switch src.(type) {
	case int:
		d.Duration = durationpb.New(time.Duration(src.(int)))
		return nil
	case int64:
		d.Duration = durationpb.New(time.Duration(src.(int64)))
		return nil
	default:
		return fmt.Errorf("can not convert %v to Duration", src)
	}
}

func (d *Duration) GetDuration() *durationpb.Duration {
	if d == nil || d.Duration == nil {
		return nil
	}
	return d.Duration
}
