package recurrence

import (
	"fmt"
	"time"

	"github.com/shindakun/attodo/internal/models"
)

// CalculateNextOccurrence calculates the next occurrence date for a recurring task
func CalculateNextOccurrence(rt *models.RecurringTask, fromDate time.Time) (*time.Time, error) {
	// Use last generated date if available, otherwise use fromDate
	baseDate := fromDate
	if rt.LastGeneratedAt != nil {
		baseDate = *rt.LastGeneratedAt
	}

	var nextDate time.Time

	switch rt.Frequency {
	case "daily":
		nextDate = baseDate.AddDate(0, 0, rt.Interval)

	case "weekly":
		nextDate = calculateNextWeekly(baseDate, rt.Interval, rt.DaysOfWeek)

	case "monthly":
		nextDate = calculateNextMonthly(baseDate, rt.Interval, rt.DayOfMonth)

	case "yearly":
		nextDate = baseDate.AddDate(rt.Interval, 0, 0)

	default:
		return nil, fmt.Errorf("unsupported frequency: %s", rt.Frequency)
	}

	// Check if we've exceeded end date
	if rt.EndDate != nil && nextDate.After(*rt.EndDate) {
		return nil, nil // No more occurrences
	}

	// Check if we've exceeded max occurrences
	if rt.MaxOccurrences > 0 && rt.OccurrenceCount >= rt.MaxOccurrences {
		return nil, nil // No more occurrences
	}

	return &nextDate, nil
}

// calculateNextWeekly calculates the next weekly occurrence
// daysOfWeek is an array of weekday integers where 0=Sunday, 1=Monday, etc.
func calculateNextWeekly(baseDate time.Time, interval int, daysOfWeek []int) time.Time {
	if len(daysOfWeek) == 0 {
		// If no specific days, just add interval weeks
		return baseDate.AddDate(0, 0, interval*7)
	}

	// Find the next matching day of week
	currentWeekday := int(baseDate.Weekday())

	// First, try to find a day in the current week
	for _, day := range daysOfWeek {
		if day > currentWeekday {
			daysToAdd := day - currentWeekday
			return baseDate.AddDate(0, 0, daysToAdd)
		}
	}

	// No day found in current week, move to next interval and take first day
	weeksToAdd := interval
	daysIntoWeek := daysOfWeek[0] // Take the first day of week from the pattern
	totalDays := (weeksToAdd * 7) - currentWeekday + daysIntoWeek

	return baseDate.AddDate(0, 0, totalDays)
}

// calculateNextMonthly calculates the next monthly occurrence
func calculateNextMonthly(baseDate time.Time, interval int, dayOfMonth int) time.Time {
	// If dayOfMonth is 0, use the same day of month as baseDate
	targetDay := dayOfMonth
	if targetDay == 0 {
		targetDay = baseDate.Day()
	}

	// Add interval months
	nextMonth := baseDate.AddDate(0, interval, 0)

	// Set to the target day, handling month-end cases
	year, month, _ := nextMonth.Date()

	// Get the last day of the target month
	lastDayOfMonth := time.Date(year, month+1, 0, 0, 0, 0, 0, nextMonth.Location()).Day()

	// If target day exceeds month length, use last day of month
	if targetDay > lastDayOfMonth {
		targetDay = lastDayOfMonth
	}

	// Preserve time from base date
	hour, min, sec := baseDate.Clock()
	return time.Date(year, month, targetDay, hour, min, sec, 0, nextMonth.Location())
}

// CalculateInitialOccurrence calculates the first occurrence date based on task creation
func CalculateInitialOccurrence(rt *models.RecurringTask, creationDate time.Time, dueDate *time.Time) *time.Time {
	// If the task has a due date, use it as the base for recurrence
	if dueDate != nil {
		return dueDate
	}

	// Otherwise, calculate from creation date
	nextOccurrence, _ := CalculateNextOccurrence(rt, creationDate)
	return nextOccurrence
}
