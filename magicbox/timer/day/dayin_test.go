package day_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/day"
)

func TestNewDayIn(t *testing.T) {
	dayIn, err := day.NewDayIn(1, 2, 3)
	assert.NoError(t, err)
	assert.NotNil(t, dayIn)
	assert.True(t, dayIn.Match(time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)))
	assert.False(t, dayIn.Match(time.Date(2023, 1, 4, 0, 0, 0, 0, time.UTC)))
	assert.True(t, dayIn.Match(time.Date(2023, 1, 2, 23, 59, 59, 999999999, time.UTC)))
	assert.True(t, dayIn.Match(time.Date(2023, 1, 2, 23, 59, 59, 1000000000, time.UTC)))
}
