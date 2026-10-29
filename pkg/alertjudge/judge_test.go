package alertjudge

import (
	"testing"
	"time"
)

func TestMatchForRequiresContinuousPoints(t *testing.T) {
	base := time.Date(2026, 10, 29, 8, 0, 0, 0, time.UTC)
	level := Level{Condition: ConditionEQ, Sample: SampleFor, Threshold: "0", ForSeconds: 60}
	points := []Point{
		{Value: 0, At: base, Raw: "0"},
		{Value: 0, At: base.Add(30 * time.Second), Raw: "0"},
		{Value: 0, At: base.Add(60 * time.Second), Raw: "0"},
	}
	if _, _, ok := Match(level, points, 30); !ok {
		t.Fatal("continuous points should hit")
	}
	broken := []Point{
		{Value: 0, At: base, Raw: "0"},
		{Value: 0, At: base.Add(120 * time.Second), Raw: "0"},
		{Value: 0, At: base.Add(180 * time.Second), Raw: "0"},
	}
	if _, _, ok := Match(level, broken, 30); ok {
		t.Fatal("gap larger than one step should not hit")
	}
	short := points[:2]
	if _, _, ok := Match(level, short, 30); ok {
		t.Fatal("span shorter than for seconds should not hit")
	}
}

func TestMatchMaxMin(t *testing.T) {
	base := time.Now()
	points := []Point{{Value: 1200, At: base}, {Value: 1800, At: base.Add(time.Second)}}
	maxLevel := Level{Condition: ConditionGT, Sample: SampleMax, Threshold: "1500"}
	if value, _, ok := Match(maxLevel, points, 0); !ok || value != 1800 {
		t.Fatalf("max should hit with 1800, got %v %v", value, ok)
	}
	minLevel := Level{Condition: ConditionLT, Sample: SampleMin, Threshold: "1500"}
	if value, _, ok := Match(minLevel, points, 0); !ok || value != 1200 {
		t.Fatalf("min should hit with 1200, got %v %v", value, ok)
	}
}
