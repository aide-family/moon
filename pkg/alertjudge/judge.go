// Package alertjudge 是告警级别条件的比较实现。查询规则（megrez）与探测规则
// （alioth）共用这份代码，避免两边对「持续 / 最大值 / 最小值」的理解分叉。
package alertjudge

import (
	"math"
	"strconv"
	"strings"
	"time"
)

// 采样方式。
const (
	SampleFor = "FOR"
	SampleMax = "MAX"
	SampleMin = "MIN"
)

// 比较条件。
const (
	ConditionEQ    = "EQ"
	ConditionNE    = "NE"
	ConditionGT    = "GT"
	ConditionGTE   = "GTE"
	ConditionLT    = "LT"
	ConditionLTE   = "LTE"
	ConditionIN    = "IN"
	ConditionNotIn = "NOT_IN"
)

// Level 是一条级别条件。
type Level struct {
	Condition  string
	Sample     string
	Threshold  string
	ForSeconds int
}

// Point 是一个采样点。
type Point struct {
	Value float64
	At    time.Time
	// Raw 是原始数值文本，用于事件展示。
	Raw string
}

// Match 按级别条件比较采样点。step 是数据源时间桶秒数，只有按桶返回的数据才非 0：
// FOR 必须保证桶连续，否则窗口中间缺失的桶会被首末点跨度掩盖，形成误告警。
func Match(level Level, points []Point, step int64) (float64, string, bool) {
	if len(points) == 0 {
		return 0, "", false
	}
	switch level.Sample {
	case SampleMax:
		best := points[0]
		for _, point := range points[1:] {
			if point.Value > best.Value {
				best = point
			}
		}
		if Compare(level.Condition, best.Value, level.Threshold) {
			return best.Value, best.Raw, true
		}
	case SampleMin:
		best := points[0]
		for _, point := range points[1:] {
			if point.Value < best.Value {
				best = point
			}
		}
		if Compare(level.Condition, best.Value, level.Threshold) {
			return best.Value, best.Raw, true
		}
	case SampleFor:
		for _, point := range points {
			if !Compare(level.Condition, point.Value, level.Threshold) {
				return 0, "", false
			}
		}
		if step > 0 {
			// 允许半个步长的容差：上游可能把桶时间戳编码成浮点秒或毫秒。
			gap := time.Duration(step) * time.Second
			for index := 1; index < len(points); index++ {
				if points[index].At.Sub(points[index-1].At) > gap+gap/2 {
					return 0, "", false
				}
			}
		}
		span := points[len(points)-1].At.Sub(points[0].At)
		if level.ForSeconds > 0 && span < time.Duration(level.ForSeconds)*time.Second {
			return 0, "", false
		}
		last := points[len(points)-1]
		return last.Value, last.Raw, true
	}
	return 0, "", false
}

// Compare 比较数值与阈值。阈值不是合法数字时一律不命中。
func Compare(condition string, value float64, threshold string) bool {
	switch condition {
	case ConditionEQ:
		n, ok := Atof(threshold)
		return ok && value == n
	case ConditionNE:
		n, ok := Atof(threshold)
		return ok && value != n
	case ConditionGT:
		n, ok := Atof(threshold)
		return ok && value > n
	case ConditionGTE:
		n, ok := Atof(threshold)
		return ok && value >= n
	case ConditionLT:
		n, ok := Atof(threshold)
		return ok && value < n
	case ConditionLTE:
		n, ok := Atof(threshold)
		return ok && value <= n
	case ConditionIN:
		return inList(value, threshold)
	case ConditionNotIn:
		return !inList(value, threshold)
	default:
		return false
	}
}

// Atof 解析阈值。
func Atof(value string) (float64, bool) {
	n, err := strconv.ParseFloat(strings.TrimSpace(value), 64)
	if err != nil || math.IsNaN(n) {
		return 0, false
	}
	return n, true
}

func inList(value float64, threshold string) bool {
	for _, part := range strings.Split(threshold, ",") {
		n, ok := Atof(strings.TrimSpace(part))
		if ok && value == n {
			return true
		}
	}
	return false
}
