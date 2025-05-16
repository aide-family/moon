package timer

import (
	"time"

	"github.com/aide-family/moon/pkg/merr"
)

func ValidateHourRange(rule []int) error {
	if len(rule) != 2 {
		return merr.ErrorParamsError("invalid hour range: %v", rule)
	}

	start := rule[0]
	end := rule[1]
	if start < 0 || start > 23 || end < 0 || end > 23 {
		return merr.ErrorParamsError("invalid hour range: %d-%d", start, end)
	}
	return nil
}

func ValidateHour(rule []int) error {
	for _, hour := range rule {
		if hour < 0 || hour > 23 {
			return merr.ErrorParamsError("invalid hour: %d", hour)
		}
	}
	return nil
}

func ValidateHourMinuteRange(rule []int) error {
	if len(rule) != 4 {
		return merr.ErrorParamsError("invalid hour minute range: %v", rule)
	}

	startHour := rule[0]
	startMinute := rule[1]
	endHour := rule[2]
	endMinute := rule[3]
	if startHour < 0 || startHour > 23 || startMinute < 0 || startMinute > 59 || endHour < 0 || endHour > 23 || endMinute < 0 || endMinute > 59 {
		return merr.ErrorParamsError("invalid hour minute range: %d:%d-%d:%d", startHour, startMinute, endHour, endMinute)
	}
	return nil
}

func ValidateDaysOfWeek(rule []int) error {
	for _, day := range rule {
		if day < int(time.Sunday) || day > int(time.Saturday) {
			return merr.ErrorParamsError("invalid days of week: %v", rule)
		}
	}
	return nil
}

func ValidateDayOfMonth(rule []int) error {
	if len(rule) != 2 {
		return merr.ErrorParamsError("invalid day of months: %v", rule)
	}
	start := rule[0]
	end := rule[1]
	if start < 1 || start > 31 || end < 1 || end > 31 {
		return merr.ErrorParamsError("invalid day of months: %v", rule)
	}
	return nil
}

func ValidateMonth(rule []int) error {
	if len(rule) != 2 {
		return merr.ErrorParamsError("invalid month: %v", rule)
	}
	start := rule[0]
	end := rule[1]
	if start < 1 || start > 12 || end < 1 || end > 12 {
		return merr.ErrorParamsError("invalid month: %v", rule)
	}
	return nil
}

func ValidateHourMinute(hour, minute int) error {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return merr.ErrorParamsError("invalid hour minute: %d-%d", hour, minute)
	}
	return nil
}
