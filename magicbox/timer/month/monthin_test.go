package month_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aide-family/magicbox/timer/month"
)

func TestNewMonthIn(t *testing.T) {
	monthIn, err := month.NewMonthIn(1, 2, 3)
	assert.Nil(t, err)
	assert.NotNil(t, monthIn)
	assert.True(t, monthIn.Match(time.Date(2023, 1, 1, 0, 0, 0, 0, time.Local)))
	assert.True(t, monthIn.Match(time.Date(2023, 2, 1, 0, 0, 0, 0, time.Local)))
	assert.True(t, monthIn.Match(time.Date(2023, 3, 1, 0, 0, 0, 0, time.Local)))
	assert.False(t, monthIn.Match(time.Date(2023, 4, 1, 0, 0, 0, 0, time.Local)))
	assert.False(t, monthIn.Match(time.Date(2023, 12, 1, 0, 0, 0, 0, time.Local)))
}
