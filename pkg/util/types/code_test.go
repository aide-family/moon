package types

import (
	"testing"
)

func TestMatchStatusCodes(t *testing.T) {
	type args struct {
		patterns   string
		statusCode int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "test1",
			args: args{
				patterns:   "2xx,3xx,4xx,5xx",
				statusCode: 200,
			},
			want: true,
		},
		{
			name: "test2",
			args: args{
				patterns:   "2xx,3xx,4xx,5xx",
				statusCode: 201,
			},
			want: true,
		},
		{
			name: "test3",
			args: args{
				patterns:   "2xx,3xx,4xx,5xx",
				statusCode: 301,
			},
			want: true,
		},
		{
			name: "test4",
			args: args{
				patterns:   "2xx,3xx,4xx,5xx",
				statusCode: 404,
			},
			want: true,
		},
		{
			name: "test5",
			args: args{
				patterns:   "2xx,3xx,4xx,5xx",
				statusCode: 500,
			},
			want: true,
		},
		{
			name: "test6",
			args: args{
				patterns:   "2xx,3xx,5xx",
				statusCode: 400,
			},
			want: false,
		},
		{
			name: "test7",
			args: args{
				patterns:   "3xx,4xx,5xx",
				statusCode: 200,
			},
			want: false,
		},
		{
			name: "test8",
			args: args{
				patterns:   "2xx,4xx,5xx",
				statusCode: 300,
			},
			want: false,
		},
		{
			name: "test9",
			args: args{
				patterns:   "",
				statusCode: 300,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchStatusCodes(tt.args.patterns, tt.args.statusCode); got != tt.want {
				t.Errorf("MatchStatusCodes() = %v, want %v", got, tt.want)
			}
		})
	}
}
