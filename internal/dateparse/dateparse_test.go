package dateparse

import (
	"testing"
	"time"
)

func TestParseNumericDate(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		input         string
		expectedDay   int
		expectedMonth time.Month
		expectedYear  int
		shouldFind    bool
	}{
		{
			name:          "MM/DD format future date",
			input:         "11/26 meeting with team",
			expectedDay:   26,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "MM/DD format past date rolls to next year",
			input:         "01/15 review document",
			expectedDay:   15,
			expectedMonth: time.January,
			expectedYear:  2025,
			shouldFind:    true,
		},
		{
			name:          "MM/DD/YYYY format",
			input:         "12/25/2024 christmas party",
			expectedDay:   25,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "MM/DD/YY format",
			input:         "06/15/25 summer event",
			expectedDay:   15,
			expectedMonth: time.June,
			expectedYear:  2025,
			shouldFind:    true,
		},
		{
			name:       "Invalid month",
			input:      "13/01 invalid date",
			shouldFind: false,
		},
		{
			name:       "Invalid day",
			input:      "11/32 invalid date",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
				if result.DueDate.Year() != tt.expectedYear {
					t.Errorf("Expected year %d, got %d", tt.expectedYear, result.DueDate.Year())
				}

				// Check that date was removed from title
				if result.CleanedTitle == tt.input {
					t.Errorf("Title was not cleaned: %s", result.CleanedTitle)
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}

func TestParseDayName(t *testing.T) {
	// Set reference to a Wednesday
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC) // Wednesday

	tests := []struct {
		name          string
		input         string
		expectedDay   int
		expectedMonth time.Month
		shouldFind    bool
	}{
		{
			name:          "friday same week",
			input:         "friday meeting",
			expectedDay:   22,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "next monday",
			input:         "next monday review",
			expectedDay:   25,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "monday (upcoming)",
			input:         "monday deadline",
			expectedDay:   25,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "saturday",
			input:         "saturday task",
			expectedDay:   23,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:       "invalid day name",
			input:      "someday task",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}

func TestParseRelativeDate(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		input         string
		expectedDay   int
		expectedMonth time.Month
		shouldFind    bool
	}{
		{
			name:          "today",
			input:         "today finish report",
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "tomorrow",
			input:         "tomorrow call client",
			expectedDay:   21,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "tmr abbreviation",
			input:         "tmr meeting",
			expectedDay:   21,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "three days",
			input:         "three days task-title",
			expectedDay:   23,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "3 days",
			input:         "3 days task-title",
			expectedDay:   23,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in 3 days",
			input:         "in 3 days call back",
			expectedDay:   23,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "5 days",
			input:         "5 days review",
			expectedDay:   25,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "1 week",
			input:         "1 week project deadline",
			expectedDay:   27,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "2 weeks",
			input:         "2 weeks vacation",
			expectedDay:   4,
			expectedMonth: time.December,
			shouldFind:    true,
		},
		{
			name:          "in 1 week",
			input:         "in 1 week presentation",
			expectedDay:   27,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in three days",
			input:         "this is due in three days",
			expectedDay:   23,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "five days",
			input:         "five days meeting",
			expectedDay:   25,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "a week",
			input:         "a week project",
			expectedDay:   27,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in a week",
			input:         "in a week deadline",
			expectedDay:   27,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "two weeks",
			input:         "two weeks sprint",
			expectedDay:   4,
			expectedMonth: time.December,
			shouldFind:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}

func TestParseMonthName(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		input         string
		expectedDay   int
		expectedMonth time.Month
		expectedYear  int
		shouldFind    bool
	}{
		{
			name:          "Nov 26 format",
			input:         "Nov 26 presentation",
			expectedDay:   26,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "26 Nov format",
			input:         "26 Nov deadline",
			expectedDay:   26,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "December 25 format",
			input:         "December 25 party",
			expectedDay:   25,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "Past date rolls to next year",
			input:         "Jan 15 resolution",
			expectedDay:   15,
			expectedMonth: time.January,
			expectedYear:  2025,
			shouldFind:    true,
		},
		{
			name:       "Invalid day",
			input:      "Nov 32 invalid",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
				if result.DueDate.Year() != tt.expectedYear {
					t.Errorf("Expected year %d, got %d", tt.expectedYear, result.DueDate.Year())
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}

func TestParseCleansTitle(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		input         string
		expectedTitle string
	}{
		{
			name:          "Date at beginning",
			input:         "11/26 meeting with team",
			expectedTitle: "meeting with team",
		},
		{
			name:          "Date in middle",
			input:         "Schedule tomorrow the review",
			expectedTitle: "Schedule the review",
		},
		{
			name:          "Date at end",
			input:         "Important deadline friday",
			expectedTitle: "Important deadline",
		},
		{
			name:          "No date",
			input:         "Regular task without date",
			expectedTitle: "Regular task without date",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if result.CleanedTitle != tt.expectedTitle {
				t.Errorf("Expected title '%s', got '%s'", tt.expectedTitle, result.CleanedTitle)
			}
		})
	}
}

func TestParseNoDate(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC)

	inputs := []string{
		"Regular task",
		"Meeting with client",
		"Review document",
		"Task with numbers 123 456",
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			result := Parse(input, refTime)

			if result.DueDate != nil {
				t.Errorf("Expected no date for '%s', but found %v", input, result.DueDate)
			}

			if result.CleanedTitle != input {
				t.Errorf("Title should remain unchanged: expected '%s', got '%s'", input, result.CleanedTitle)
			}
		})
	}
}

func TestParseAdvancedRelative(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC) // Wednesday, Nov 20, 2024

	tests := []struct {
		name          string
		input         string
		expectedDay   int
		expectedMonth time.Month
		expectedYear  int
		shouldFind    bool
	}{
		{
			name:          "3 days from now",
			input:         "3 days from now finish report",
			expectedDay:   23,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "2 weeks from now",
			input:         "2 weeks from now vacation",
			expectedDay:   4,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "1 month from now",
			input:         "1 month from now review",
			expectedDay:   20,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "3 months from now",
			input:         "3 months from now project deadline",
			expectedDay:   20,
			expectedMonth: time.February,
			expectedYear:  2025,
			shouldFind:    true,
		},
		{
			name:          "in 2 months",
			input:         "in 2 months quarterly review",
			expectedDay:   20,
			expectedMonth: time.January,
			expectedYear:  2025,
			shouldFind:    true,
		},
		{
			name:          "end of month",
			input:         "end of month report",
			expectedDay:   30,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "end of the month",
			input:         "end of the month summary",
			expectedDay:   30,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "end of week",
			input:         "end of week cleanup",
			expectedDay:   24, // Sunday
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "end of the week",
			input:         "end of the week status",
			expectedDay:   24, // Sunday
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "next quarter",
			input:         "next quarter planning",
			expectedDay:   1,
			expectedMonth: time.January,
			expectedYear:  2025,
			shouldFind:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
				if result.DueDate.Year() != tt.expectedYear {
					t.Errorf("Expected year %d, got %d", tt.expectedYear, result.DueDate.Year())
				}

				// Check that date was removed from title
				if result.CleanedTitle == tt.input {
					t.Errorf("Title was not cleaned: %s", result.CleanedTitle)
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}

func TestParseTime(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name          string
		input         string
		expectedDay   int
		expectedMonth time.Month
		expectedHour  int
		expectedMin   int
		shouldFind    bool
	}{
		{
			name:          "tomorrow at 3pm",
			input:         "tomorrow at 3pm meeting",
			expectedDay:   21,
			expectedMonth: time.November,
			expectedHour:  15,
			expectedMin:   0,
			shouldFind:    true,
		},
		{
			name:          "tomorrow at 3:30pm",
			input:         "tomorrow at 3:30pm doctor appointment",
			expectedDay:   21,
			expectedMonth: time.November,
			expectedHour:  15,
			expectedMin:   30,
			shouldFind:    true,
		},
		{
			name:          "11/26 at 2pm",
			input:         "11/26 at 2pm team meeting",
			expectedDay:   26,
			expectedMonth: time.November,
			expectedHour:  14,
			expectedMin:   0,
			shouldFind:    true,
		},
		{
			name:          "friday 5:30pm",
			input:         "friday 5:30pm happy hour",
			expectedDay:   22,
			expectedMonth: time.November,
			expectedHour:  17,
			expectedMin:   30,
			shouldFind:    true,
		},
		{
			name:          "today at 9am",
			input:         "today at 9am standup",
			expectedDay:   20,
			expectedMonth: time.November,
			expectedHour:  9,
			expectedMin:   0,
			shouldFind:    true,
		},
		{
			name:          "tomorrow at 12pm (noon)",
			input:         "tomorrow at 12pm lunch",
			expectedDay:   21,
			expectedMonth: time.November,
			expectedHour:  12,
			expectedMin:   0,
			shouldFind:    true,
		},
		{
			name:          "tomorrow at 12am (midnight)",
			input:         "tomorrow at 12am release",
			expectedDay:   21,
			expectedMonth: time.November,
			expectedHour:  0,
			expectedMin:   0,
			shouldFind:    true,
		},
		{
			name:          "next monday at 10:30am",
			input:         "next monday at 10:30am project review",
			expectedDay:   25,
			expectedMonth: time.November,
			expectedHour:  10,
			expectedMin:   30,
			shouldFind:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
				if result.DueDate.Hour() != tt.expectedHour {
					t.Errorf("Expected hour %d, got %d", tt.expectedHour, result.DueDate.Hour())
				}
				if result.DueDate.Minute() != tt.expectedMin {
					t.Errorf("Expected minute %d, got %d", tt.expectedMin, result.DueDate.Minute())
				}

				// Check that date/time was removed from title
				if result.CleanedTitle == tt.input {
					t.Errorf("Title was not cleaned: %s", result.CleanedTitle)
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}

func TestParseTimeBasedRelative(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 14, 30, 0, 0, time.UTC) // Wednesday, Nov 20, 2024 at 2:30 PM

	tests := []struct {
		name          string
		input         string
		expectedHour  int
		expectedMin   int
		expectedDay   int
		expectedMonth time.Month
		shouldFind    bool
	}{
		{
			name:          "in 2 hours",
			input:         "this is due in 2 hours",
			expectedHour:  16,
			expectedMin:   30,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in two hours",
			input:         "this is due in two hours",
			expectedHour:  16,
			expectedMin:   30,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "2 hours",
			input:         "meeting 2 hours from now",
			expectedHour:  16,
			expectedMin:   30,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in 30 minutes",
			input:         "call back in 30 minutes",
			expectedHour:  15,
			expectedMin:   0,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in thirty minutes",
			input:         "reminder in thirty minutes",
			expectedHour:  15,
			expectedMin:   0,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in 90 minutes",
			input:         "appointment in 90 minutes",
			expectedHour:  16,
			expectedMin:   0,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in one hour",
			input:         "meeting in one hour",
			expectedHour:  15,
			expectedMin:   30,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in an hour",
			input:         "call in an hour",
			expectedHour:  15,
			expectedMin:   30,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "in fifteen minutes",
			input:         "break in fifteen minutes",
			expectedHour:  14,
			expectedMin:   45,
			expectedDay:   20,
			expectedMonth: time.November,
			shouldFind:    true,
		},
		{
			name:          "3 hours (crosses day boundary)",
			input:         "review in 10 hours",
			expectedHour:  0,
			expectedMin:   30,
			expectedDay:   21, // Next day
			expectedMonth: time.November,
			shouldFind:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
				if result.DueDate.Hour() != tt.expectedHour {
					t.Errorf("Expected hour %d, got %d", tt.expectedHour, result.DueDate.Hour())
				}
				if result.DueDate.Minute() != tt.expectedMin {
					t.Errorf("Expected minute %d, got %d", tt.expectedMin, result.DueDate.Minute())
				}

				// Check that time text was removed from title
				if result.CleanedTitle == tt.input {
					t.Errorf("Title was not cleaned: %s", result.CleanedTitle)
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}

func TestParseAdditionalNaturalLanguage(t *testing.T) {
	refTime := time.Date(2024, 11, 20, 12, 0, 0, 0, time.UTC) // Wednesday, Nov 20, 2024

	tests := []struct {
		name          string
		input         string
		expectedDay   int
		expectedMonth time.Month
		expectedYear  int
		shouldFind    bool
	}{
		{
			name:          "next month",
			input:         "next month review",
			expectedDay:   20,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "next year",
			input:         "next year planning",
			expectedDay:   20,
			expectedMonth: time.November,
			expectedYear:  2025,
			shouldFind:    true,
		},
		{
			name:          "start of month",
			input:         "start of month report",
			expectedDay:   1,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "beginning of the month",
			input:         "beginning of the month check-in",
			expectedDay:   1,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "mid month",
			input:         "mid month review",
			expectedDay:   15,
			expectedMonth: time.December,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "start of week",
			input:         "start of week standup",
			expectedDay:   25,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
		{
			name:          "beginning of the week",
			input:         "beginning of the week planning",
			expectedDay:   25,
			expectedMonth: time.November,
			expectedYear:  2024,
			shouldFind:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse(tt.input, refTime)

			if tt.shouldFind {
				if result.DueDate == nil {
					t.Errorf("Expected to find date, but got nil")
					return
				}

				if result.DueDate.Day() != tt.expectedDay {
					t.Errorf("Expected day %d, got %d", tt.expectedDay, result.DueDate.Day())
				}
				if result.DueDate.Month() != tt.expectedMonth {
					t.Errorf("Expected month %v, got %v", tt.expectedMonth, result.DueDate.Month())
				}
				if result.DueDate.Year() != tt.expectedYear {
					t.Errorf("Expected year %d, got %d", tt.expectedYear, result.DueDate.Year())
				}

				// Check that date was removed from title
				if result.CleanedTitle == tt.input {
					t.Errorf("Title was not cleaned: %s", result.CleanedTitle)
				}
			} else {
				if result.DueDate != nil {
					t.Errorf("Expected no date, but found %v", result.DueDate)
				}
			}
		})
	}
}
