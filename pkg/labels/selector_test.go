package labels

import "testing"

func TestRequirements_Matches(t *testing.T) {
	type args struct {
		labels Labels
	}
	tests := []struct {
		name string
		x    Requirements
		args args
		want bool
	}{
		{
			name: "empty require",
			x:    Requirements{},
			args: args{
				labels: Set{},
			},
			want: true,
		},
		{
			name: "empty labels",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar"},
				},
			},
			args: args{
				labels: Set{},
			},
			want: false,
		},
		{
			name: "match with single value",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar"},
				},
			},
			args: args{
				labels: Set{
					"foo": "bar",
				},
			},
			want: true,
		},
		{
			name: "not match with single value",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"foo"},
				},
			},
			args: args{
				labels: Set{
					"foo": "bar",
				},
			},
			want: false,
		},
		{
			name: "match with mutil keys and values",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar"},
				},
				{
					key:      "a",
					operator: Equals,
					values:   []string{"a"},
				},
			},
			args: args{
				labels: Set{
					"foo": "bar",
					"a":   "a",
				},
			},
			want: true,
		},
		{
			name: "not match with mutil keys and values",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar"},
				},
				{
					key:      "a",
					operator: Equals,
					values:   []string{"a"},
				},
			},
			args: args{
				labels: Set{
					"foo": "bar",
					"a":   "b",
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.Matches(tt.args.labels); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequirements_String(t *testing.T) {
	tests := []struct {
		name string
		x    Requirements
		want string
	}{
		{
			name: "empty",
			x:    Requirements{},
			want: "",
		},
		{
			name: "single",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar"},
				},
			},
			want: "foo in (bar)",
		},
		{
			name: "mutil",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar"},
				},
				{
					key:      "a",
					operator: Equals,
					values:   []string{"a"},
				},
			},
			want: "foo in (bar), a = a",
		},
		{
			name: "not",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar", "barr"},
				},
				{
					key:      "a",
					operator: NotIn,
					values:   []string{"a"},
				},
			},
			want: "foo in (bar,barr), a notin (a)",
		},
		{
			name: "not exist",
			x: Requirements{
				{
					key:      "foo",
					operator: In,
					values:   []string{"bar"},
				},
				{
					key:      "a",
					operator: NotExist,
					values:   []string{},
				},
			},
			want: "foo in (bar), !a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.x.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
