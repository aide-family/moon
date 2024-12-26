package types

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"google.golang.org/protobuf/types/known/durationpb"
)

var _ json.Unmarshaler = (*Time)(nil)

// Time 包装后的时间类型
type Time struct {
	time.Time
}

// MarshalJSON 实现 json.Marshaler 接口
func (t *Time) MarshalJSON() ([]byte, error) {
	// 返回字符串形式的时间
	return TextJoinToBytes(`"`, t.String(), `"`), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (t *Time) UnmarshalJSON(data []byte) error {
	// 去掉字符串中的引号
	s := strings.Trim(string(data), `"`)
	// 解析时间字符串
	tt, err := time.ParseInLocation(time.DateTime, s, time.Local)
	if err != nil {
		return err
	}
	// 将解析后的时间赋值给 Time
	t.Time = tt
	return nil
}

// String 字符串
func (t *Time) String() string {
	if t == nil {
		return ""
	}
	return t.Time.Format(time.DateTime)
}

// Unix 时间戳
func (t *Time) Unix() int64 {
	if t == nil {
		return 0
	}
	return t.Time.Unix()
}

// NewTime 创建一个 Time
func NewTime(t time.Time) *Time {
	return &Time{
		Time: t,
	}
}

// NewTimeByString 从字符串创建一个 Time
func NewTimeByString(s string, layout ...string) *Time {
	lay := time.DateTime
	if len(layout) > 0 {
		lay = layout[0]
	}
	t, err := time.ParseInLocation(lay, s, time.Local)
	if err != nil {
		return NewTime(time.Now())
	}
	return NewTime(t)
}

// NewTimeByUnix 从unix 创建一个 Time
func NewTimeByUnix(unix int64) *Time {
	t := time.Unix(unix, 0)
	return NewTime(t)
}

// Scan 现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (t *Time) Scan(value interface{}) error {
	switch s := value.(type) {
	case time.Time:
		t.Time = s
	case string:
		tt, err := time.ParseInLocation(time.DateTime, s, time.Local)
		if err != nil {
			return err
		}
		t.Time = tt
	case nil:
		t.Time = time.Time{}
	default:
		return fmt.Errorf("can not convert %v to Time", value)
	}
	return nil
}

// Value 实现 driver.Valuer 接口，Value
func (t *Time) Value() (driver.Value, error) {
	return t.Time, nil
}

var (
	_ driver.Valuer = (*Duration)(nil)
	_ sql.Scanner   = (*Duration)(nil)
)

// NewDuration 创建一个 Duration
func NewDuration(dur *durationpb.Duration) *Duration {
	return &Duration{
		Duration: dur,
	}
}

// Duration 包装后的时间类型
type Duration struct {
	Duration *durationpb.Duration
}

// String 字符串
func (d *Duration) String() string {
	return d.GetDuration().String()
}

// MarshalJSON 实现 json.Marshaler 接口
func (d *Duration) MarshalJSON() ([]byte, error) {
	// 返回字符串形式的时间
	return TextJoinToBytes(`"`, d.String(), `"`), nil
}

// UnmarshalJSON 实现 json.Unmarshaler 接口
func (d *Duration) UnmarshalJSON(data []byte) error {
	// 去掉字符串中的引号
	s := strings.Trim(string(data), `"`)
	// 解析时间字符串
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	// 将解析后的时间赋值给 Time
	d.Duration = durationpb.New(dur)
	return nil
}

// CronTime 定义一个时间间隔，单位为秒
func (d *Duration) CronTime() string {
	seconds := d.GetDuration().AsDuration().Seconds()
	if seconds < 10 {
		seconds = 10
	}
	return TextJoin("@every ", strconv.Itoa(int(seconds)), "s")
}

// Value 实现 driver.Valuer 接口，Value
func (d *Duration) Value() (driver.Value, error) {
	return int64(d.GetDuration().AsDuration()), nil
}

// Scan 现 sql.Scanner 接口，Scan 将 value 扫描至 Jsonb
func (d *Duration) Scan(src any) error {
	switch s := src.(type) {
	case int:
		d.Duration = durationpb.New(time.Duration(s))
		return nil
	case int64:
		d.Duration = durationpb.New(time.Duration(s))
		return nil
	default:
		return fmt.Errorf("can not convert %v to Duration", src)
	}
}

// GetDuration 获取 Duration
func (d *Duration) GetDuration() *durationpb.Duration {
	if d == nil || d.Duration == nil {
		return nil
	}
	return d.Duration
}
