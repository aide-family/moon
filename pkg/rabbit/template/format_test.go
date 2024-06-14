package template

import (
	"bytes"
	"testing"
)

func TestParser(t *testing.T) {
	type args struct {
		format string
		in     any
	}
	tests := []struct {
		name    string
		args    args
		wantOut string
		wantErr bool
	}{
		// =============================== err case ==================================
		{
			name: "format empty",
			args: args{
				format: "",
			},
			wantOut: "",
			wantErr: true,
		},
		{
			name: "data nil",
			args: args{
				format: "",
			},
			wantOut: "",
			wantErr: true,
		},
		// =============================== normal case ==================================
		{
			name: "full filed",
			args: args{
				format: "{{ $labels.instance }} 的值大于 {{ $value }}",
				in: map[string]interface{}{
					"value": 999,
					"labels": map[string]string{
						"instance": "test",
					},
				},
			},
			wantOut: "test 的值大于 999",
			wantErr: false,
		},
		{
			name: "missing fields",
			args: args{
				format: "{{ $labels.instance }} 的值大于 {{ $value }}",
				in: map[string]interface{}{
					"value":  999,
					"labels": map[string]string{},
				},
			},
			wantOut: "<no value> 的值大于 999",
			wantErr: false,
		},
		{
			name: "empty fields",
			args: args{
				format: "{{ $labels.instance }} 的值大于 {{ $value }}",
				in: map[string]interface{}{
					"value": 999,
					"labels": map[string]string{
						"instance": "",
					},
				},
			},
			wantOut: " 的值大于 999",
			wantErr: false,
		},
		{
			name: "redundant fields",
			args: args{
				format: "{{ $labels.instance }} 的值大于 {{ $value }}",
				in: map[string]interface{}{
					"value":  999,
					"values": 9998,
					"labels": map[string]string{
						"instance":  "test",
						"instances": "test",
					},
				},
			},
			wantOut: "test 的值大于 999",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := &bytes.Buffer{}
			err := Parser(tt.args.format, tt.args.in, out)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotOut := out.String(); gotOut != tt.wantOut {
				t.Errorf("Parser() gotOut = %v, want %v", gotOut, tt.wantOut)
			}
		})
	}
}

func Test_replaceString(t *testing.T) {
	type args struct {
		str string
	}
	tests := []struct {
		name  string
		args  args
		wantS string
	}{
		{
			name: "empty case",
			args: args{
				str: "",
			},
			wantS: "",
		},
		{
			name: "no filed need replace case 1",
			args: args{
				str: "this is test string",
			},
			wantS: "this is test string",
		},
		{
			name: "no filed need replace case 2",
			args: args{
				str: "{{ .labels.instance }} 的值大于 {{ .value }}",
			},
			wantS: "{{ .labels.instance }} 的值大于 {{ .value }}",
		},
		{
			name: "normal case",
			args: args{
				str: "{{ $labels.instance }} 的值大于 {{ $value }}",
			},
			wantS: "{{ .labels.instance }} 的值大于 {{ .value }}",
		},
		{
			name: "normal case 2",
			args: args{
				str: "{{ $labels.instance }} 的值大于 {{ .value }}",
			},
			wantS: "{{ .labels.instance }} 的值大于 {{ .value }}",
		},
		{
			name: "normal case 3",
			args: args{
				str: "{{ $labels$instance }} 的值大于 {{ .value }}",
			},
			wantS: "{{ .labels.instance }} 的值大于 {{ .value }}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotS := replaceString(tt.args.str); gotS != tt.wantS {
				t.Errorf("replaceString() = %v, want %v", gotS, tt.wantS)
			}
		})
	}
}
