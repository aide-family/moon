package data

import "testing"

func TestParsePromInstantValue(t *testing.T) {
	cases := []struct {
		name    string
		payload string
		want    float64
		found   bool
		wantErr bool
	}{
		{
			name:    "vector",
			payload: `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"instance":"a"},"value":[1700000000,"42.5"]}]}}`,
			want:    42.5,
			found:   true,
		},
		{
			name:    "vector numeric value",
			payload: `{"status":"success","data":{"resultType":"vector","result":[{"metric":{},"value":[1700000000,7]}]}}`,
			want:    7,
			found:   true,
		},
		{
			name:    "scalar",
			payload: `{"status":"success","data":{"resultType":"scalar","result":[1700000000,"0.25"]}}`,
			want:    0.25,
			found:   true,
		},
		{
			name:    "empty vector",
			payload: `{"status":"success","data":{"resultType":"vector","result":[]}}`,
			found:   false,
		},
		{
			name:    "error status",
			payload: `{"status":"error","error":"bad query"}`,
			wantErr: true,
		},
		{
			name:    "invalid json",
			payload: `not-json`,
			wantErr: true,
		},
		{
			name:    "malformed sample",
			payload: `{"status":"success","data":{"resultType":"vector","result":[{"value":[1700000000]}]}}`,
			wantErr: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			value, found, err := parsePromInstantValue(tc.payload)
			if tc.wantErr {
				if err == nil {
					t.Fatal("want error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if found != tc.found {
				t.Fatalf("found = %v, want %v", found, tc.found)
			}
			if found && value != tc.want {
				t.Fatalf("value = %v, want %v", value, tc.want)
			}
		})
	}
}
