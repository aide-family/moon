package oncall

import (
	"testing"
	"time"
)

func mustCompile(t *testing.T, specs ...Spec) []Window {
	t.Helper()
	windows, err := Compile(specs)
	if err != nil {
		t.Fatalf("compile windows: %v", err)
	}
	return windows
}

func TestNormalize(t *testing.T) {
	normalized, err := Normalize([]Spec{
		{Weekdays: []string{"friday", "MONDAY", "Monday"}, Start: "9:00", End: "18:00"},
		{Weekdays: []string{"MONDAY", "FRIDAY"}, Start: "09:00", End: "18:00"},
	})
	if err != nil {
		t.Fatalf("normalize: %v", err)
	}
	if len(normalized) != 1 {
		t.Fatalf("duplicate windows not merged: %+v", normalized)
	}
	if got := normalized[0]; got.Start != "09:00" || got.End != "18:00" || len(got.Weekdays) != 2 || got.Weekdays[0] != "MONDAY" || got.Weekdays[1] != "FRIDAY" {
		t.Fatalf("unexpected normalized window: %+v", got)
	}

	invalid := [][]Spec{
		{{Weekdays: []string{}, Start: "09:00", End: "18:00"}},
		{{Weekdays: []string{"FUNDAY"}, Start: "09:00", End: "18:00"}},
		{{Weekdays: []string{"MONDAY"}, Start: "09:0", End: "18:00"}},
		{{Weekdays: []string{"MONDAY"}, Start: "24:00", End: "18:00"}},
		{{Weekdays: []string{"MONDAY"}, Start: "09:00", End: "09:00"}},
	}
	for _, specs := range invalid {
		if _, err := Normalize(specs); err == nil {
			t.Fatalf("invalid windows accepted: %+v", specs)
		}
	}
}

func TestActive(t *testing.T) {
	weekdays := mustCompile(t, Spec{Weekdays: []string{"MONDAY", "TUESDAY", "WEDNESDAY", "THURSDAY", "FRIDAY"}, Start: "09:00", End: "18:00"})
	cases := []struct {
		name string
		at   time.Time
		want bool
	}{
		{name: "monday start", at: time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC), want: true},
		{name: "monday middle", at: time.Date(2026, 10, 26, 12, 30, 0, 0, time.UTC), want: true},
		{name: "monday end is exclusive", at: time.Date(2026, 10, 26, 18, 0, 0, 0, time.UTC), want: false},
		{name: "saturday", at: time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC), want: false},
		{name: "before start", at: time.Date(2026, 10, 26, 8, 59, 0, 0, time.UTC), want: false},
	}
	for _, item := range cases {
		if got := Active(weekdays, item.at); got != item.want {
			t.Fatalf("%s: Active = %v, want %v", item.name, got, item.want)
		}
	}
	if !Active(nil, time.Now()) {
		t.Fatal("empty windows must always be active")
	}
}

func TestActiveAcrossMidnight(t *testing.T) {
	windows := mustCompile(t, Spec{Weekdays: []string{"MONDAY"}, Start: "22:00", End: "06:00"})
	cases := []struct {
		name string
		at   time.Time
		want bool
	}{
		{name: "monday evening", at: time.Date(2026, 10, 26, 23, 0, 0, 0, time.UTC), want: true},
		{name: "tuesday early", at: time.Date(2026, 10, 27, 2, 0, 0, 0, time.UTC), want: true},
		{name: "tuesday end is exclusive", at: time.Date(2026, 10, 27, 6, 0, 0, 0, time.UTC), want: false},
		{name: "monday early belongs to sunday", at: time.Date(2026, 10, 26, 2, 0, 0, 0, time.UTC), want: false},
		{name: "tuesday evening belongs to tuesday", at: time.Date(2026, 10, 27, 23, 0, 0, 0, time.UTC), want: false},
	}
	for _, item := range cases {
		if got := Active(windows, item.at); got != item.want {
			t.Fatalf("%s: Active = %v, want %v", item.name, got, item.want)
		}
	}
}

func TestActiveUsesProvidedLocation(t *testing.T) {
	windows := mustCompile(t, Spec{Weekdays: []string{"MONDAY"}, Start: "09:00", End: "10:00"})
	location, err := LoadLocation("Asia/Shanghai")
	if err != nil {
		t.Fatal(err)
	}
	// 2026-10-26 01:30 UTC 是上海时间周一 09:30。
	at := time.Date(2026, 10, 26, 1, 30, 0, 0, time.UTC).In(location)
	if !Active(windows, at) {
		t.Fatal("window should be active in Asia/Shanghai")
	}
	if Active(windows, at.UTC()) {
		t.Fatal("same instant in UTC must not be treated as local time")
	}
}

func TestMarshalRoundTrip(t *testing.T) {
	raw, err := Marshal([]Spec{{Weekdays: []string{"SUNDAY"}, Start: "08:00", End: "20:00"}})
	if err != nil {
		t.Fatal(err)
	}
	specs, err := Unmarshal(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(specs) != 1 || specs[0].Weekdays[0] != "SUNDAY" || specs[0].Start != "08:00" {
		t.Fatalf("round trip = %+v", specs)
	}
	for _, empty := range []string{"", "null", "[]"} {
		specs, err := Unmarshal(empty)
		if err != nil || len(specs) != 0 {
			t.Fatalf("unmarshal %q = %+v, %v", empty, specs, err)
		}
	}
	if _, err := Unmarshal("{bad"); err == nil {
		t.Fatal("invalid json accepted")
	}
}

func TestNormalizeTimezone(t *testing.T) {
	name, err := NormalizeTimezone("")
	if err != nil || name != DefaultTimezone {
		t.Fatalf("default timezone = %q, %v", name, err)
	}
	if _, err := NormalizeTimezone("Mars/Olympus"); err == nil {
		t.Fatal("invalid timezone accepted")
	}
}
