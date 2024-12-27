package vobj

import "testing"

func TestStringJudge(t *testing.T) {
	tests := []struct {
		condition EventCondition
		data      string
		threshold string
		expected  bool
	}{
		// EQ tests
		{EventConditionEQ, "10", "10", true},
		{EventConditionEQ, "10", "20", false},

		// NE tests
		{EventConditionNE, "10", "20", true},
		{EventConditionNE, "10", "10", false},

		// GT tests
		{EventConditionGT, "20", "10", true},
		{EventConditionGT, "10", "20", false},

		// GTE tests
		{EventConditionGTE, "20", "20", true},
		{EventConditionGTE, "10", "20", false},

		// LT tests
		{EventConditionLT, "10", "20", true},
		{EventConditionLT, "20", "10", false},

		// LTE tests
		{EventConditionLTE, "10", "10", true},
		{EventConditionLTE, "20", "10", false},

		// Contain tests
		{EventConditionContain, "hello world", "world", true},
		{EventConditionContain, "hello world", "earth", false},

		// Prefix tests
		{EventConditionPrefix, "hello world", "hello", true},
		{EventConditionPrefix, "hello world", "world", false},

		// Suffix tests
		{EventConditionSuffix, "hello world", "world", true},
		{EventConditionSuffix, "hello world", "hello", false},

		// Regular expression tests
		{EventConditionRegular, "hello123", `\d+`, true}, // matches numbers
		{EventConditionRegular, "hello", `\d+`, false},   // doesn't match numbers
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.condition)), func(t *testing.T) {
			actual := tt.condition.stringJudge(tt.data, tt.threshold)
			if actual != tt.expected {
				t.Errorf("stringJudge(%v, %v) = %v; want %v", tt.data, tt.threshold, actual, tt.expected)
			}
		})
	}
}
