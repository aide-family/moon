package types

import (
	"reflect"
	"testing"
	"time"
)

func TestNewTimeEngine(t *testing.T) {
	type args struct {
		opts []Option
	}
	ts, _ := time.Parse(time.DateTime, "2024-12-18 00:00:00")
	tests := []struct {
		name string
		args args
		want bool
		ts   time.Time
	}{
		{
			name: "默认状态下为true",
			args: args{opts: []Option{}},
			want: true,
			ts:   ts,
		},
		{
			name: "指定小时时间段为true",
			args: args{opts: []Option{WithConfiguration(&HourRange{
				Start: 0,
				End:   1,
			})}},
			want: true,
			ts:   ts,
		},
		{
			name: "指定小时时间段为false",
			args: args{opts: []Option{WithConfiguration(&HourRange{
				Start: 1,
				End:   12,
			})}},
			want: false,
			ts:   ts,
		},
		{
			name: "指定星期为true",
			args: args{opts: []Option{WithConfiguration(&DaysOfWeek{1, 2, 3})}},
			want: true,
			ts:   ts,
		},
		{
			name: "指定星期为false",
			args: args{opts: []Option{WithConfiguration(&DaysOfWeek{1, 2, 4})}},
			want: false,
			ts:   ts,
		},
		{
			name: "指定日期为true",
			args: args{opts: []Option{WithConfiguration(&DaysOfMonth{Start: 10, End: 20})}},
			want: true,
			ts:   ts,
		},
		{
			name: "指定日期为false",
			args: args{opts: []Option{WithConfiguration(&DaysOfMonth{Start: 1, End: 10})}},
			want: false,
			ts:   ts,
		},
		{
			name: "指定月份为true",
			args: args{opts: []Option{WithConfiguration(&Months{Start: 11, End: 12})}},
			want: true,
			ts:   ts,
		},
		{
			name: "指定月份为false",
			args: args{opts: []Option{WithConfiguration(&Months{Start: 1, End: 10})}},
			want: false,
			ts:   ts,
		},
		{
			name: "所有条件共存为true",
			args: args{opts: []Option{
				WithConfiguration(&DaysOfWeek{1, 2, 3}),
				WithConfiguration(&HourRange{Start: 0, End: 1}),
				WithConfiguration(&DaysOfMonth{Start: 10, End: 20}),
				WithConfiguration(&Months{Start: 11, End: 12}),
			}},
			want: true,
			ts:   ts,
		},
		{
			name: "所有条件共存为false",
			args: args{opts: []Option{
				WithConfiguration(&DaysOfWeek{1, 2, 3}),
				WithConfiguration(&HourRange{Start: 1, End: 12}),
				WithConfiguration(&DaysOfMonth{Start: 1, End: 10}),
				WithConfiguration(&Months{Start: 1, End: 10}),
			}},
			want: false,
			ts:   ts,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewTimeEngine(tt.args.opts...).IsAllowed(tt.ts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewTimeEngine() = %v, want %v", got, tt.want)
			}
		})
	}
}
