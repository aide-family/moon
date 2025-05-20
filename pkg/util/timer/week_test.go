package timer_test

import (
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/util/timer"
	"github.com/aide-family/moon/pkg/util/validate"
)

func TestNewDaysOfWeek_ValidInput_ReturnsMatcher(t *testing.T) {
	rule := []int{0, 2, 4, 6} // Sunday, Tuesday, Thursday, Saturday
	matcher, err := timer.NewDaysOfWeek(rule)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if validate.IsNil(matcher) {
		t.Errorf("Expected a non-nil matcher")
	}
}

func TestNewDaysOfWeek_InvalidInput_ReturnsError(t *testing.T) {
	rule := []int{0, 7, 10} // Invalid days
	_, err := timer.NewDaysOfWeek(rule)
	if err == nil {
		t.Errorf("Expected an error, got nil")
	}
	t.Errorf("Expected a ParamsError, got %v", err)
}

func TestNewDaysOfWeek_EmptyInput_ReturnsMatcher(t *testing.T) {
	rule := []int{}
	matcher, err := timer.NewDaysOfWeek(rule)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if validate.IsNil(matcher) {
		t.Errorf("Expected a non-nil matcher")
	}
}

func TestDaysOfWeek_Match_ValidDays(t *testing.T) {
	rule := []int{0, 2, 4, 6} // Sunday, Tuesday, Thursday, Saturday
	matcher, _ := timer.NewDaysOfWeek(rule)

	tests := []struct {
		date     time.Time
		expected bool
	}{
		{time.Date(2023, 10, 1, 0, 0, 0, 0, time.UTC), true},  // Sunday
		{time.Date(2023, 10, 3, 0, 0, 0, 0, time.UTC), true},  // Tuesday
		{time.Date(2023, 10, 5, 0, 0, 0, 0, time.UTC), true},  // Thursday
		{time.Date(2023, 10, 7, 0, 0, 0, 0, time.UTC), true},  // Saturday
		{time.Date(2023, 10, 2, 0, 0, 0, 0, time.UTC), false}, // Monday
		{time.Date(2023, 10, 4, 0, 0, 0, 0, time.UTC), false}, // Wednesday
		{time.Date(2023, 10, 6, 0, 0, 0, 0, time.UTC), false}, // Friday
	}

	for _, test := range tests {
		result := matcher.Match(test.date)
		if result != test.expected {
			t.Errorf("Expected %v, got %v for date %v", test.expected, result, test.date)
		}
	}
}
