package event

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetRealtimeTableNames(t *testing.T) {
	tests := []struct {
		name     string
		teamID   uint32
		start    time.Time
		end      time.Time
		expected []string
	}{
		{
			name:   "single week",
			teamID: 1,
			start:  time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC), // Monday
			end:    time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC), // Wednesday
			expected: []string{
				"team_realtime_alerts_1_20240304",
			},
		},
		{
			name:   "multiple weeks",
			teamID: 1,
			start:  time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC),  // Monday
			end:    time.Date(2024, 3, 18, 0, 0, 0, 0, time.UTC), // Monday
			expected: []string{
				"team_realtime_alerts_1_20240304",
				"team_realtime_alerts_1_20240311",
				"team_realtime_alerts_1_20240318",
			},
		},
		{
			name:   "start not on monday",
			teamID: 1,
			start:  time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC),  // Wednesday
			end:    time.Date(2024, 3, 20, 0, 0, 0, 0, time.UTC), // Wednesday
			expected: []string{
				"team_realtime_alerts_1_20240304",
				"team_realtime_alerts_1_20240311",
				"team_realtime_alerts_1_20240318",
			},
		},
		{
			name:     "invalid time range",
			teamID:   1,
			start:    time.Date(2024, 3, 6, 0, 0, 0, 0, time.UTC),
			end:      time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC),
			expected: nil,
		},
		{
			name:   "single day",
			teamID: 1,
			start:  time.Date(2024, 3, 4, 0, 0, 0, 0, time.UTC),
			end:    time.Date(2024, 3, 4, 23, 59, 59, 0, time.UTC),
			expected: []string{
				"team_realtime_alerts_1_20240304",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetRealtimeTableNames(tt.teamID, tt.start, tt.end, nil)
			assert.Equal(t, tt.expected, result)
		})
	}
}
