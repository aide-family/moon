package timer_test

import (
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/util/timer"
)

func TestNewHourRange_InvalidLength_ErrorReturned(t *testing.T) {
	rule := []int{1}
	_, err := timer.NewHourRange(rule)
	if err == nil {
		t.Errorf("Expected error for invalid length, got nil")
	}
}

func TestNewHourRange_InvalidHour_ErrorReturned(t *testing.T) {
	rule := []int{24, 25}
	_, err := timer.NewHourRange(rule)
	if err == nil {
		t.Errorf("Expected error for invalid hour, got nil")
	}
}

func TestNewHourRange_ValidInput_MatcherCreated(t *testing.T) {
	rule := []int{10, 15}
	matcher, err := timer.NewHourRange(rule)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if matcher == nil {
		t.Errorf("Expected matcher to be created, got nil")
	}
}

func TestHourRangeMatch_HourInRange_ReturnsTrue(t *testing.T) {
	rule := []int{10, 15}
	matcher, _ := timer.NewHourRange(rule)
	tm := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	if !matcher.Match(tm) {
		t.Errorf("Expected true for hour in range, got false")
	}
}

func TestHourRangeMatch_HourNotInRange_ReturnsFalse(t *testing.T) {
	rule := []int{10, 15}
	matcher, _ := timer.NewHourRange(rule)
	tm := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)
	if matcher.Match(tm) {
		t.Errorf("Expected false for hour not in range, got true")
	}
}

func TestNewHour_ValidInput_MatcherCreated(t *testing.T) {
	rule := []int{10, 15}
	matcher, err := timer.NewHour(rule)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if matcher == nil {
		t.Errorf("Expected matcher to be created, got nil")
	}
}

func TestHourMatch_HourInList_ReturnsTrue(t *testing.T) {
	rule := []int{10, 15}
	matcher, _ := timer.NewHour(rule)
	tm := time.Date(2023, 10, 1, 10, 0, 0, 0, time.UTC)
	if !matcher.Match(tm) {
		t.Errorf("Expected true for hour in list, got false")
	}
}

func TestHourMatch_HourNotInList_ReturnsFalse(t *testing.T) {
	rule := []int{10, 15}
	matcher, _ := timer.NewHour(rule)
	tm := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)
	if matcher.Match(tm) {
		t.Errorf("Expected false for hour not in list, got true")
	}
}

func TestNewHourMinuteRange_ValidInput_MatcherCreated(t *testing.T) {
	start := timer.HourMinute{Hour: 10, Minute: 30}
	end := timer.HourMinute{Hour: 15, Minute: 45}
	matcher, err := timer.NewHourMinuteRange(start, end)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if matcher == nil {
		t.Errorf("Expected matcher to be created, got nil")
	}
}

func TestHourMinuteRangeMatch_TimeInRange_ReturnsTrue(t *testing.T) {
	start := timer.HourMinute{Hour: 10, Minute: 30}
	end := timer.HourMinute{Hour: 15, Minute: 45}
	matcher, _ := timer.NewHourMinuteRange(start, end)
	tm := time.Date(2023, 10, 1, 12, 0, 0, 0, time.UTC)
	if !matcher.Match(tm) {
		t.Errorf("Expected true for time in range, got false")
	}
}

func TestHourMinuteRangeMatch_TimeNotInRange_ReturnsFalse(t *testing.T) {
	start := timer.HourMinute{Hour: 10, Minute: 30}
	end := timer.HourMinute{Hour: 15, Minute: 45}
	matcher, _ := timer.NewHourMinuteRange(start, end)
	tm := time.Date(2023, 10, 1, 9, 0, 0, 0, time.UTC)
	if matcher.Match(tm) {
		t.Errorf("Expected false for time not in range, got true")
	}
}

func TestHourMinuteRangeMatch_CrossMidnight_ReturnsTrue(t *testing.T) {
	start := timer.HourMinute{Hour: 23, Minute: 30}
	end := timer.HourMinute{Hour: 1, Minute: 45}
	matcher, _ := timer.NewHourMinuteRange(start, end)
	tm := time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC)
	if !matcher.Match(tm) {
		t.Errorf("Expected true for time in cross midnight range, got false")
	}
}

func TestHourMinute_Gt_ReturnsCorrectResult(t *testing.T) {
	hm1 := timer.HourMinute{Hour: 10, Minute: 30}
	hm2 := timer.HourMinute{Hour: 9, Minute: 45}
	if !hm1.GT(&hm2) {
		t.Errorf("Expected true for GT, got false")
	}
}

func TestHourMinute_Lt_ReturnsCorrectResult(t *testing.T) {
	hm1 := timer.HourMinute{Hour: 10, Minute: 30}
	hm2 := timer.HourMinute{Hour: 11, Minute: 45}
	if !hm1.LT(&hm2) {
		t.Errorf("Expected true for LT, got false")
	}
}
