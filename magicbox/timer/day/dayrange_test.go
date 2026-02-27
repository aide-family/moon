package day_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/day"
)

func TestNewDayRange(t *testing.T) {
	dayRange, err := day.NewDayRange(1, 20)
	assert.Nil(t, err)
	assert.NotNil(t, dayRange)
	assert.True(t, dayRange.Match(time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)))
	assert.True(t, dayRange.Match(time.Date(2023, 1, 10, 0, 0, 0, 0, time.Local)))
	assert.True(t, dayRange.Match(time.Date(2023, 1, 20, 0, 0, 0, 0, time.Local)))
	assert.False(t, dayRange.Match(time.Date(2023, 1, 21, 0, 0, 0, 0, time.Local)))
	assert.False(t, dayRange.Match(time.Date(2023, 1, 31, 0, 0, 0, 0, time.Local)))
}
