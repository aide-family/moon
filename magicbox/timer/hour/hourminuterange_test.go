package hour_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/hour"
)

func TestNewHourMinuteRange(t *testing.T) {
	start := time.Date(2023, 1, 1, 10, 10, 0, 0, time.Local)
	end := time.Date(2023, 1, 1, 20, 20, 0, 0, time.Local)
	hourMinuteRange, err := hour.NewHourMinuteRange(start, end)
	assert.Nil(t, err)
	assert.NotNil(t, hourMinuteRange)
	assert.True(t, hourMinuteRange.Match(time.Date(2023, 1, 1, 10, 10, 0, 0, time.Local)))
	assert.True(t, hourMinuteRange.Match(time.Date(2023, 1, 1, 20, 20, 0, 0, time.Local)))
	assert.False(t, hourMinuteRange.Match(time.Date(2023, 1, 1, 9, 10, 0, 0, time.Local)))
	assert.False(t, hourMinuteRange.Match(time.Date(2023, 1, 1, 20, 21, 0, 0, time.Local)))
	assert.False(t, hourMinuteRange.Match(time.Date(2023, 1, 1, 21, 20, 0, 0, time.Local)))
}
