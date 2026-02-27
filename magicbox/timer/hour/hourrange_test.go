package hour_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/hour"
)

func TestNewHourRange(t *testing.T) {
	hourRange, err := hour.NewHourRange(10, 23)
	assert.Nil(t, err)
	assert.NotNil(t, hourRange)
	assert.True(t, hourRange.Match(time.Date(2023, 1, 1, 10, 0, 0, 0, time.UTC)))
	assert.True(t, hourRange.Match(time.Date(2023, 1, 1, 23, 0, 0, 0, time.UTC)))
	assert.False(t, hourRange.Match(time.Date(2023, 1, 1, 25, 0, 0, 0, time.UTC)))
}
