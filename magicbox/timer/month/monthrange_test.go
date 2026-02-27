package month_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/month"
)

func TestNewMonthRange(t *testing.T) {
	monthRange, err := month.NewMonthRange(2, 8)
	assert.Nil(t, err)
	assert.NotNil(t, monthRange)
	assert.True(t, monthRange.Match(time.Date(2023, 2, 1, 0, 0, 0, 0, time.Local)))
	assert.True(t, monthRange.Match(time.Date(2023, 5, 1, 0, 0, 0, 0, time.Local)))
	assert.True(t, monthRange.Match(time.Date(2023, 8, 1, 0, 0, 0, 0, time.Local)))
	assert.False(t, monthRange.Match(time.Date(2023, 9, 1, 0, 0, 0, 0, time.Local)))
	assert.False(t, monthRange.Match(time.Date(2023, 12, 1, 0, 0, 0, 0, time.Local)))
	assert.False(t, monthRange.Match(time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)))
}
