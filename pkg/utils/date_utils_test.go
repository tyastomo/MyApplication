package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCalculateWorkingDays(t *testing.T) {
	loc, _ := time.LoadLocation("UTC") // Use a consistent location for tests

	testCases := []struct {
		name          string
		startDate     time.Time
		endDate       time.Time
		expectedDays  int
		expectError   bool // For future use if validation is added
	}{
		{
			name:         "Typical Week (Mon-Fri)",
			startDate:    time.Date(2023, time.October, 2, 0, 0, 0, 0, loc), // Monday
			endDate:      time.Date(2023, time.October, 6, 0, 0, 0, 0, loc), // Friday
			expectedDays: 5,
		},
		{
			name:         "Week including weekend (Mon-Sun)",
			startDate:    time.Date(2023, time.October, 2, 0, 0, 0, 0, loc), // Monday
			endDate:      time.Date(2023, time.October, 8, 0, 0, 0, 0, loc), // Sunday
			expectedDays: 5,
		},
		{
			name:         "Starts on Saturday, ends on Sunday (only weekend)",
			startDate:    time.Date(2023, time.October, 7, 0, 0, 0, 0, loc), // Saturday
			endDate:      time.Date(2023, time.October, 8, 0, 0, 0, 0, loc), // Sunday
			expectedDays: 0,
		},
		{
			name:         "Starts on Friday, ends on Monday (Fri, Mon)",
			startDate:    time.Date(2023, time.October, 6, 0, 0, 0, 0, loc), // Friday
			endDate:      time.Date(2023, time.October, 9, 0, 0, 0, 0, loc), // Monday
			expectedDays: 2,
		},
		{
			name:         "Single Working Day (Wednesday)",
			startDate:    time.Date(2023, time.October, 4, 0, 0, 0, 0, loc), // Wednesday
			endDate:      time.Date(2023, time.October, 4, 0, 0, 0, 0, loc), // Wednesday
			expectedDays: 1,
		},
		{
			name:         "Single Weekend Day (Saturday)",
			startDate:    time.Date(2023, time.October, 7, 0, 0, 0, 0, loc), // Saturday
			endDate:      time.Date(2023, time.October, 7, 0, 0, 0, 0, loc), // Saturday
			expectedDays: 0,
		},
		{
			name:         "End date before start date",
			startDate:    time.Date(2023, time.October, 6, 0, 0, 0, 0, loc),
			endDate:      time.Date(2023, time.October, 2, 0, 0, 0, 0, loc),
			expectedDays: 0, // Or handle as error if preferred by design
		},
		{
			name:         "Two full weeks (Mon to Sun, then Mon to Sun)",
			startDate:    time.Date(2023, time.October, 2, 0, 0, 0, 0, loc),  // Monday
			endDate:      time.Date(2023, time.October, 15, 0, 0, 0, 0, loc), // Sunday
			expectedDays: 10,
		},
		{
			name:         "Across month boundary",
			startDate:    time.Date(2023, time.September, 29, 0, 0, 0, 0, loc), // Friday
			endDate:      time.Date(2023, time.October, 2, 0, 0, 0, 0, loc),    // Monday
			expectedDays: 2, // Sept 29 (Fri), Oct 2 (Mon)
		},
		{
            name:         "February non-leap year",
            startDate:    time.Date(2023, time.February, 1, 0, 0, 0, 0, loc),
            endDate:      time.Date(2023, time.February, 28, 0, 0, 0, 0, loc),
            expectedDays: 20, // 4 full weeks
        },
        {
            name:         "February leap year (2024)",
            startDate:    time.Date(2024, time.February, 1, 0, 0, 0, 0, loc),
            endDate:      time.Date(2024, time.February, 29, 0, 0, 0, 0, loc),
            expectedDays: 21, // 29 days, 8 weekend days -> 21 working days
        },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualDays := CalculateWorkingDays(tc.startDate, tc.endDate)
			assert.Equal(t, tc.expectedDays, actualDays)
		})
	}
}
