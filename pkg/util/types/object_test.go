package types

import (
	"reflect"
	"testing"
)

func TestUnwrapOr(t *testing.T) {
	type args[T any] struct {
		p        *T
		fallback []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want T
	}
	tests := []testCase[int]{
		{
			name: "nil",
			args: args[int]{
				p:        nil,
				fallback: []int{1, 2, 3},
			},
			want: 1,
		},
		{
			name: "not nil",
			args: args[int]{
				p:        Of(4),
				fallback: []int{1, 2, 3},
			},
			want: 4,
		},
		{
			name: "not nil",
			args: args[int]{
				p:        nil,
				fallback: []int{},
			},
			want: 0,
		},
		{
			name: "not nil",
			args: args[int]{
				p:        nil,
				fallback: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := UnwrapOr(tt.args.p, tt.args.fallback...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UnwrapOr() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractPointerOr(t *testing.T) {
	type args[T any] struct {
		value    any
		fallback []T
	}
	type testCase[T any] struct {
		name string
		args args[T]
		want T
	}
	tests := []testCase[int]{
		{
			name: "nil",
			args: args[int]{
				value:    nil,
				fallback: []int{1, 2, 3},
			},
			want: 1,
		},
		{
			name: "not nil",
			args: args[int]{
				value:    Of(4),
				fallback: []int{1, 2, 3},
			},
			want: 4,
		},
		{
			name: "not nil",
			args: args[int]{
				value:    nil,
				fallback: []int{},
			},
			want: 0,
		},
		{
			name: "not nil",
			args: args[int]{
				value:    nil,
				fallback: nil,
			},
			want: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ExtractPointerOr(tt.args.value, tt.args.fallback...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ExtractPointerOr() = %v, want %v", got, tt.want)
			}
		})
	}
}
