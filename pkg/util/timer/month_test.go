package timer_test

import (
	"testing"
	"time"

	"github.com/aide-family/moon/pkg/util/timer"
)

func TestNewDayOfMonths_ValidInput_ReturnsMatcher(t *testing.T) {
	rule := []int{1, 10}
	matcher, err := timer.NewDayOfMonths(rule)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if matcher == nil {
		t.Errorf("Expected non-nil matcher, got nil")
	}
}

func TestNewDayOfMonths_InvalidInput_ReturnsError(t *testing.T) {
	tests := []struct {
		rule []int
	}{
		{[]int{0, 32}},
		{[]int{1, 32}},
		{[]int{32, 31}},
		{[]int{1}},
		{[]int{1, 2, 3}},
	}

	for _, test := range tests {
		_, err := timer.NewDayOfMonths(test.rule)
		if err == nil {
			t.Errorf("Expected error for rule %v, got nil", test.rule)
		}
	}
}

func TestDayOfMonths_Match_ValidRange_ReturnsTrue(t *testing.T) {
	rule := []int{1, 10}
	matcher, _ := timer.NewDayOfMonths(rule)
	date := time.Date(2023, time.January, 5, 0, 0, 0, 0, time.UTC)
	if !matcher.Match(date) {
		t.Errorf("Expected true, got false")
	}
}

func TestDayOfMonths_Match_InvalidRange_ReturnsFalse(t *testing.T) {
	rule := []int{1, 10}
	matcher, _ := timer.NewDayOfMonths(rule)
	date := time.Date(2023, time.January, 11, 0, 0, 0, 0, time.UTC)
	if matcher.Match(date) {
		t.Errorf("Expected false, got true")
	}
}

func TestDayOfMonths_Match_StartGreaterThanEnd_ReturnsTrue(t *testing.T) {
	rule := []int{25, 5}
	matcher, _ := timer.NewDayOfMonths(rule)
	date := time.Date(2023, time.January, 30, 0, 0, 0, 0, time.UTC)
	if !matcher.Match(date) {
		t.Errorf("Expected true, got false")
	}
}

func TestNewMonth_ValidInput_ReturnsMatcher(t *testing.T) {
	rule := []int{1, 6}
	matcher, err := timer.NewMonth(rule)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if matcher == nil {
		t.Errorf("Expected non-nil matcher, got nil")
	}
}

func TestNewMonth_InvalidInput_ReturnsError(t *testing.T) {
	tests := []struct {
		rule []int
	}{
		{[]int{0, 13}},
		{[]int{1, 13}},
		{[]int{13, 1}},
		{[]int{1}},
		{[]int{1, 2, 3}},
	}

	for _, test := range tests {
		_, err := timer.NewMonth(test.rule)
		if err == nil {
			t.Errorf("Expected error for rule %v, got nil", test.rule)
		}
	}
}

func TestMonth_Match_ValidRange_ReturnsTrue(t *testing.T) {
	rule := []int{1, 6}
	matcher, _ := timer.NewMonth(rule)
	date := time.Date(2023, time.March, 1, 0, 0, 0, 0, time.UTC)
	if !matcher.Match(date) {
		t.Errorf("Expected true, got false")
	}
}

func TestMonth_Match_InvalidRange_ReturnsFalse(t *testing.T) {
	rule := []int{1, 6}
	matcher, _ := timer.NewMonth(rule)
	date := time.Date(2023, time.July, 1, 0, 0, 0, 0, time.UTC)
	if matcher.Match(date) {
		t.Errorf("Expected false, got true")
	}
}

func TestMonth_Match_StartGreaterThanEnd_ReturnsTrue(t *testing.T) {
	rule := []int{11, 3}
	matcher, _ := timer.NewMonth(rule)
	date := time.Date(2023, time.December, 1, 0, 0, 0, 0, time.UTC)
	if !matcher.Match(date) {
		t.Errorf("Expected true, got false")
	}
}
