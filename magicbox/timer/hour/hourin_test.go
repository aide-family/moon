package hour_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/hour"
)

func TestNewHourIn(t *testing.T) {
	hourIn, err := hour.NewHourIn(1, 2, 3, 4)
	assert.Nil(t, err)
	assert.NotNil(t, hourIn)
	assert.True(t, hourIn.Match(time.Date(2023, 1, 1, 1, 0, 0, 0, time.Local)))
	assert.True(t, hourIn.Match(time.Date(2023, 1, 1, 4, 0, 0, 0, time.Local)))
	assert.False(t, hourIn.Match(time.Date(2023, 1, 1, 5, 0, 0, 0, time.Local)))
	assert.False(t, hourIn.Match(time.Date(2023, 1, 1, 23, 0, 0, 0, time.Local)))
}
