package week_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/week"
)

func TestNewWeekIn(t *testing.T) {
	weekIn, err := week.NewWeekIn(time.Monday, time.Sunday)
	assert.Nil(t, err)
	assert.NotNil(t, weekIn)
	assert.True(t, weekIn.Match(time.Date(2025, 8, 3, 0, 0, 0, 0, time.UTC)))
	assert.True(t, weekIn.Match(time.Date(2025, 8, 4, 0, 0, 0, 0, time.UTC)))
	assert.False(t, weekIn.Match(time.Date(2025, 8, 5, 0, 0, 0, 0, time.UTC)))
	assert.False(t, weekIn.Match(time.Date(2025, 8, 6, 0, 0, 0, 0, time.UTC)))
}
