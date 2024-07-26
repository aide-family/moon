package labels

import (
	"testing"
)

func TestNewRequirement(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		op      Operator
		vals    []string
		wantErr bool
	}{
		{
			name:    "Empty key",
			key:     "",
			op:      Equals,
			vals:    []string{"value"},
			wantErr: true,
		},
		{
			name:    "Empty operator",
			key:     "key",
			op:      Operator(""),
			vals:    []string{"value"},
			wantErr: true,
		},
		{
			name:    "Invalid operator",
			key:     "key",
			op:      Operator("xxxx"),
			vals:    []string{"value"},
			wantErr: true,
		},
		{
			name:    "Exists operator with values",
			key:     "key",
			op:      Exists,
			vals:    []string{"value"},
			wantErr: true,
		},
		{
			name:    "NotExist operator without values",
			key:     "key",
			op:      NotExist,
			vals:    []string{},
			wantErr: false,
		},
		{
			name:    "Equals operator with one value",
			key:     "key",
			op:      Equals,
			vals:    []string{"value"},
			wantErr: false,
		},
		{
			name:    "Equals operator with multiple values",
			key:     "key",
			op:      Equals,
			vals:    []string{"value1", "value2"},
			wantErr: true,
		},
		{
			name:    "In operator with values",
			key:     "key",
			op:      In,
			vals:    []string{"value1", "value2"},
			wantErr: false,
		},
		{
			name:    "In operator without values",
			key:     "key",
			op:      In,
			vals:    []string{},
			wantErr: true,
		},
		{
			name:    "GreaterThan operator with valid value",
			key:     "key",
			op:      GreaterThan,
			vals:    []string{"10"},
			wantErr: false,
		},
		{
			name:    "GreaterThan operator with invalid value",
			key:     "key",
			op:      GreaterThan,
			vals:    []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "LessThan operator with valid value",
			key:     "key",
			op:      LessThan,
			vals:    []string{"10"},
			wantErr: false,
		},
		{
			name:    "LessThan operator with invalid value",
			key:     "key",
			op:      LessThan,
			vals:    []string{"invalid"},
			wantErr: true,
		},
		{
			name:    "GreaterOrEqual operator with valid value",
			key:     "key",
			op:      GreaterOrEqual,
			vals:    []string{"10"},
			wantErr: false,
		},
		{
			name:    "LessOrEqual operator with valid value",
			key:     "key",
			op:      LessOrEqual,
			vals:    []string{"10"},
			wantErr: false,
		},
		{
			name:    "NotIn operator with values",
			key:     "key",
			op:      NotIn,
			vals:    []string{"value1", "value2"},
			wantErr: false,
		},
		{
			name:    "NotIn operator without values",
			key:     "key",
			op:      NotIn,
			vals:    []string{},
			wantErr: true,
		},
		{
			name:    "NotEquals operator with values",
			key:     "key",
			op:      NotEquals,
			vals:    []string{"value1", "value2"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRequirement(tt.key, tt.op, tt.vals)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRequirement() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got == nil {
				t.Errorf("NewRequirement() returned nil for a valid input")
			}
		})
	}
}

func TestRequirement_Matches(t *testing.T) {
	type fields struct {
		key      string
		operator Operator
		values   []string
	}
	type args struct {
		labels Labels
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "Matching label",
			fields: fields{
				key:      "key",
				operator: Equals,
				values:   []string{"value"},
			},
			args: args{
				labels: Set{"key": "value"},
			},
			want: true,
		},
		{
			name: "Non-matching label",
			fields: fields{
				key:      "key",
				operator: Equals,
				values:   []string{"value"},
			},
			args: args{
				labels: Set{"key": "other"},
			},
			want: false,
		},
		{
			name: "Matching label with multiple values",
			fields: fields{
				key:      "key",
				operator: In,
				values:   []string{"value1", "value2"},
			},
			args: args{
				labels: Set{"key": "value1"},
			},
			want: true,
		},
		{
			name: "Non-matching label with multiple values",
			fields: fields{
				key:      "key",
				operator: In,
				values:   []string{"value1", "value2"},
			},
			args: args{
				labels: Set{"key": "other"},
			},
			want: false,
		},
		{
			name: "Matching label with greater than operator",
			fields: fields{
				key:      "key",
				operator: GreaterThan,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "11"},
			},
			want: true,
		},
		{
			name: "Non-matching label with greater than operator",
			fields: fields{
				key:      "key",
				operator: GreaterThan,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "9"},
			},
			want: false,
		},
		{
			name: "Matching label with greater than or equal operator",
			fields: fields{
				key:      "key",
				operator: GreaterOrEqual,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "10"},
			},
			want: true,
		},
		{
			name: "Non-matching label with greater than or equal operator",
			fields: fields{
				key:      "key",
				operator: GreaterOrEqual,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "9"},
			},
			want: false,
		},
		{
			name: "Matching label with less than operator",
			fields: fields{
				key:      "key",
				operator: LessThan,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "9"},
			},
			want: true,
		},
		{
			name: "Non-matching label with less than operator",
			fields: fields{
				key:      "key",
				operator: LessThan,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "11"},
			},
			want: false,
		},
		{
			name: "Matching label with less than or equal operator",
			fields: fields{
				key:      "key",
				operator: LessOrEqual,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "10"},
			},
			want: true,
		},
		{
			name: "Non-matching label with less than or equal operator",
			fields: fields{
				key:      "key",
				operator: LessOrEqual,
				values:   []string{"10"},
			},
			args: args{
				labels: Set{"key": "11"},
			},
			want: false,
		},
		{
			name: "Matching label with not equals operator",
			fields: fields{
				key:      "key",
				operator: NotEquals,
				values:   []string{"value"},
			},
			args: args{
				labels: Set{"key": "other"},
			},
			want: true,
		},
		{
			name: "Non-matching label with not equals operator",
			fields: fields{
				key:      "key",
				operator: NotEquals,
				values:   []string{"value"},
			},
			args: args{
				labels: Set{"key": "value"},
			},
			want: false,
		},
		{
			name: "Matching label with not exists operator",
			fields: fields{
				key:      "key",
				operator: NotExist,
				values:   []string{},
			},
			args: args{
				labels: Set{"other": "value"},
			},
			want: true,
		},
		{
			name: "Non-matching label with not exists operator",
			fields: fields{
				key:      "key",
				operator: NotExist,
				values:   []string{},
			},
			args: args{
				labels: Set{"key": "value"},
			},
			want: false,
		},
		{
			name: "Matching label with not in operator",
			fields: fields{
				key:      "key",
				operator: NotIn,
				values:   []string{"value1", "value2"},
			},
			args: args{
				labels: Set{"key": "other"},
			},
			want: true,
		},
		{
			name: "Non-matching label with not in operator",
			fields: fields{
				key:      "key",
				operator: NotIn,
				values:   []string{"value1", "value2"},
			},
			args: args{
				labels: Set{"key": "value1"},
			},
			want: false,
		},
		{
			name: "Matching label with exists operator",
			fields: fields{
				key:      "key",
				operator: Exists,
				values:   []string{},
			},
			args: args{
				labels: Set{"key": "value"},
			},
			want: true,
		},
		{
			name: "Non-matching label with exists operator",
			fields: fields{
				key:      "key",
				operator: Exists,
				values:   []string{},
			},
			args: args{
				labels: Set{"other": "value"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Requirement{
				key:      tt.fields.key,
				operator: tt.fields.operator,
				values:   tt.fields.values,
			}
			if got := r.Matches(tt.args.labels); got != tt.want {
				t.Errorf("Matches() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRequirement_String(t *testing.T) {
	type fields struct {
		key      string
		operator Operator
		values   []string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Equals operator",
			fields: fields{
				key:      "key",
				operator: Equals,
				values:   []string{"value"},
			},
			want: "key = value",
		},
		{
			name: "In operator",
			fields: fields{
				key:      "key",
				operator: In,
				values:   []string{"value1", "value2"},
			},
			want: "key in (value1,value2)",
		},
		{
			name: "Not equals operator",
			fields: fields{
				key:      "key",
				operator: NotEquals,
				values:   []string{"value"},
			},
			want: "key != value",
		},
		{
			name: "Not in operator",
			fields: fields{
				key:      "key",
				operator: NotIn,
				values:   []string{"value1", "value2"},
			},
			want: "key notin (value1,value2)",
		},
		{
			name: "Greater than operator",
			fields: fields{
				key:      "key",
				operator: GreaterThan,
				values:   []string{"10"},
			},
			want: "key > 10",
		},
		{
			name: "Greater than or equal operator",
			fields: fields{
				key:      "key",
				operator: GreaterOrEqual,
				values:   []string{"10"},
			},
			want: "key >= 10",
		},
		{
			name: "Less than operator",
			fields: fields{
				key:      "key",
				operator: LessThan,
				values:   []string{"10"},
			},
			want: "key < 10",
		},
		{
			name: "Less than or equal operator",
			fields: fields{
				key:      "key",
				operator: LessOrEqual,
				values:   []string{"10"},
			},
			want: "key <= 10",
		},
		{
			name: "Exists operator",
			fields: fields{
				key:      "key",
				operator: Exists,
				values:   []string{},
			},
			want: "key",
		},
		{
			name: "Not exists operator",
			fields: fields{
				key:      "key",
				operator: NotExist,
				values:   []string{},
			},
			want: "!key",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &Requirement{
				key:      tt.fields.key,
				operator: tt.fields.operator,
				values:   tt.fields.values,
			}
			if got := r.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}
