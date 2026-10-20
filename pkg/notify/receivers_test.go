package notify

import (
	"reflect"
	"testing"
)

func TestSplitReceivers(t *testing.T) {
	cases := []struct {
		raw  string
		want []string
	}{
		{"", nil},
		{" , ", nil},
		{"a@example.com", []string{"a@example.com"}},
		{" a@example.com , b@example.com ", []string{"a@example.com", "b@example.com"}},
		{"a@example.com,,b@example.com,", []string{"a@example.com", "b@example.com"}},
		{"a@example.com;b@example.com\nc@example.com", []string{"a@example.com", "b@example.com", "c@example.com"}},
	}
	for _, item := range cases {
		if got := SplitReceivers(item.raw); !reflect.DeepEqual(got, item.want) {
			t.Fatalf("SplitReceivers(%q) = %#v, want %#v", item.raw, got, item.want)
		}
	}
}
