package utils

import "time"

// CalculateWorkingDays iterates from start to end date (inclusive) and
// counts days that are not Saturday or Sunday.
func CalculateWorkingDays(start time.Time, end time.Time) int {
	// Normalize dates to the beginning of the day to avoid issues with time components
	startDate := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, start.Location())
	endDate := time.Date(end.Year(), end.Month(), end.Day(), 0, 0, 0, 0, end.Location())

	if endDate.Before(startDate) {
		return 0
	}

	workingDays := 0
	current := startDate
	for !current.After(endDate) {
		weekday := current.Weekday()
		if weekday != time.Saturday && weekday != time.Sunday {
			workingDays++
		}
		current = current.AddDate(0, 0, 1) // Move to the next day
	}
	return workingDays
}
