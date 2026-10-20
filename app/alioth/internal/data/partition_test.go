package data

import (
	"testing"
	"time"
)

func TestWeekStart(t *testing.T) {
	from, until := weekStart(time.Date(2026, 10, 21, 13, 45, 0, 0, time.UTC))
	if from.Weekday() != time.Monday || from.Hour() != 0 || from.Minute() != 0 {
		t.Fatalf("week start = %s", from)
	}
	if until.Sub(from) != 7*24*time.Hour {
		t.Fatalf("week span = %s", until.Sub(from))
	}
	if !from.Before(until) || !until.After(from) {
		t.Fatalf("invalid week range %s - %s", from, until)
	}
}

func TestPartitionIndexesCovered(t *testing.T) {
	for _, table := range partitionedTables {
		if len(partitionIndexes[table]) == 0 {
			t.Fatalf("%s has no partition index statements", table)
		}
	}
}
