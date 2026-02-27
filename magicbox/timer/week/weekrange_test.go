package week_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/week"
)

func TestNewWeekRange(t *testing.T) {
	weekRange, err := week.NewWeekRange(time.Monday, time.Friday)
	assert.Nil(t, err)
	assert.NotNil(t, weekRange)
	assert.True(t, weekRange.Match(time.Date(2025, 8, 1, 0, 0, 0, 0, time.Local)))
	assert.False(t, weekRange.Match(time.Date(2025, 8, 2, 0, 0, 0, 0, time.Local)))
	assert.False(t, weekRange.Match(time.Date(2025, 8, 3, 0, 0, 0, 0, time.Local)))
	assert.True(t, weekRange.Match(time.Date(2025, 8, 4, 0, 0, 0, 0, time.Local)))
	assert.True(t, weekRange.Match(time.Date(2025, 8, 5, 0, 0, 0, 0, time.Local)))
	assert.True(t, weekRange.Match(time.Date(2025, 8, 6, 0, 0, 0, 0, time.Local)))
	assert.True(t, weekRange.Match(time.Date(2025, 8, 7, 0, 0, 0, 0, time.Local)))
	assert.True(t, weekRange.Match(time.Date(2025, 8, 8, 0, 0, 0, 0, time.Local)))
}
