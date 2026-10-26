package data

import (
	"testing"
	"time"

	"github.com/aide-family/Alioth/app/alioth/internal/data/ent"
)

func TestMatchOnCall(t *testing.T) {
	event := notifyEvent{
		EventUID: "evt-1", LevelUID: "lv1", DatasourceUID: "ds1",
		Labels: `{"env":"prod","instance":"10.0.0.1:9100"}`,
	}
	tests := []struct {
		name   string
		onCall *ent.NotificationOnCall
		event  notifyEvent
		want   bool
	}{
		{
			name:   "empty conditions match any event",
			onCall: &ent.NotificationOnCall{Labels: "{}"},
			event:  event,
			want:   true,
		},
		{
			name: "all conditions match",
			onCall: &ent.NotificationOnCall{
				Labels: `{"env":"prod"}`, LevelUids: []string{"lv1", "lv2"}, DatasourceUids: []string{"ds1"},
			},
			event: event,
			want:  true,
		},
		{
			name:   "level mismatch",
			onCall: &ent.NotificationOnCall{Labels: "{}", LevelUids: []string{"lv2"}},
			event:  event,
			want:   false,
		},
		{
			name:   "datasource mismatch",
			onCall: &ent.NotificationOnCall{Labels: "{}", DatasourceUids: []string{"ds2"}},
			event:  event,
			want:   false,
		},
		{
			name:   "labels mismatch",
			onCall: &ent.NotificationOnCall{Labels: `{"env":"dev"}`},
			event:  event,
			want:   false,
		},
		{
			name:   "label regex matches whole value",
			onCall: &ent.NotificationOnCall{Labels: `{"instance":"10\\.0\\..*"}`},
			event:  event,
			want:   true,
		},
		{
			name:   "empty label value requires the key",
			onCall: &ent.NotificationOnCall{Labels: `{"team":""}`},
			event:  event,
			want:   false,
		},
	}
	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			if got := matchOnCall(item.onCall, item.event); got != item.want {
				t.Fatalf("matchOnCall = %v, want %v", got, item.want)
			}
		})
	}
}

func TestOnCallWindowActive(t *testing.T) {
	business := `[{"weekdays":["MONDAY","TUESDAY","WEDNESDAY","THURSDAY","FRIDAY"],"start":"09:00","end":"18:00"}]`
	overnight := `[{"weekdays":["MONDAY"],"start":"22:00","end":"06:00"}]`
	tests := []struct {
		name       string
		timeWindow string
		at         time.Time
		want       bool
	}{
		{name: "inside business hours", timeWindow: business, at: time.Date(2026, 10, 26, 9, 0, 0, 0, time.UTC), want: true},
		{name: "before start", timeWindow: business, at: time.Date(2026, 10, 26, 8, 59, 0, 0, time.UTC), want: false},
		{name: "end is exclusive", timeWindow: business, at: time.Date(2026, 10, 26, 18, 0, 0, 0, time.UTC), want: false},
		{name: "weekend", timeWindow: business, at: time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC), want: false},
		{name: "overnight evening", timeWindow: overnight, at: time.Date(2026, 10, 26, 23, 0, 0, 0, time.UTC), want: true},
		{name: "overnight next morning", timeWindow: overnight, at: time.Date(2026, 10, 27, 2, 0, 0, 0, time.UTC), want: true},
		{name: "overnight boundary is exclusive", timeWindow: overnight, at: time.Date(2026, 10, 27, 6, 0, 0, 0, time.UTC), want: false},
		{name: "empty windows mean all day", timeWindow: "[]", at: time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC), want: true},
		{name: "empty string means all day", timeWindow: "", at: time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC), want: true},
		{name: "invalid stored json skips binding", timeWindow: "{bad", at: time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC), want: false},
	}
	for _, item := range tests {
		t.Run(item.name, func(t *testing.T) {
			link := &ent.NotificationOnCallGroup{TimeWindows: item.timeWindow}
			if got := onCallWindowActive(link, item.at); got != item.want {
				t.Fatalf("onCallWindowActive = %v, want %v", got, item.want)
			}
		})
	}
}
