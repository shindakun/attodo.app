package dateparse

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseResult contains the extracted date and cleaned title
type ParseResult struct {
	DueDate      *time.Time // Parsed due date (nil if none found)
	CleanedTitle string     // Title with date text removed
	OriginalDate string     // The original date string that was matched
}

// Parse extracts date from title and returns cleaned title
// referenceTime is used as the base for relative dates (usually time.Now())
func Parse(title string, referenceTime time.Time) ParseResult {
	result := ParseResult{
		CleanedTitle: title,
		DueDate:      nil,
		OriginalDate: "",
	}

	// Try different parsers in order of specificity
	parsers := []func(string, time.Time) (*time.Time, string, string){
		parseNumericDate,
		parseDayName,
		parseRelativeDate,
		parseMonthName,
	}

	for _, parser := range parsers {
		if date, original, cleaned := parser(title, referenceTime); date != nil {
			result.DueDate = date
			result.OriginalDate = original
			result.CleanedTitle = normalizeWhitespace(cleaned)

			// After finding a date, try to parse time from remaining text
			if timeVal, timeOriginal, timeCleaned := parseTime(result.CleanedTitle); timeVal != nil {
				// Apply the time to the date
				year, month, day := result.DueDate.Year(), result.DueDate.Month(), result.DueDate.Day()
				hour, min, _ := timeVal.Clock()
				*result.DueDate = time.Date(year, month, day, hour, min, 0, 0, result.DueDate.Location())
				result.OriginalDate = original + " " + timeOriginal
				result.CleanedTitle = normalizeWhitespace(timeCleaned)
			}
			break
		}
	}

	// If no date was found, check if there's a time specified (e.g., "due at 9am")
	// If so, assume today's date with that time
	if result.DueDate == nil {
		if timeVal, timeOriginal, timeCleaned := parseTime(title); timeVal != nil {
			// Use today's date with the specified time
			year, month, day := referenceTime.Year(), referenceTime.Month(), referenceTime.Day()
			hour, min, _ := timeVal.Clock()
			today := time.Date(year, month, day, hour, min, 0, 0, referenceTime.Location())
			result.DueDate = &today
			result.OriginalDate = timeOriginal
			result.CleanedTitle = normalizeWhitespace(timeCleaned)
		}
	}

	return result
}

// normalizeWhitespace removes extra whitespace and trims
func normalizeWhitespace(s string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")
	return strings.TrimSpace(s)
}

// parseNumericDate handles formats like "11/26", "11/26/24", "11/26/2024"
func parseNumericDate(text string, refTime time.Time) (*time.Time, string, string) {
	// Match MM/DD or MM/DD/YY or MM/DD/YYYY
	patterns := []struct {
		regex  *regexp.Regexp
		format string
	}{
		{
			// MM/DD/YYYY or MM/DD/YY
			regex:  regexp.MustCompile(`\b(\d{1,2})/(\d{1,2})/(\d{2,4})\b`),
			format: "date_with_year",
		},
		{
			// MM/DD (no year)
			regex:  regexp.MustCompile(`\b(\d{1,2})/(\d{1,2})\b`),
			format: "date_no_year",
		},
	}

	for _, p := range patterns {
		matches := p.regex.FindStringSubmatch(text)
		if len(matches) == 0 {
			continue
		}

		original := matches[0]
		month, _ := strconv.Atoi(matches[1])
		day, _ := strconv.Atoi(matches[2])

		// Validate month and day
		if month < 1 || month > 12 || day < 1 || day > 31 {
			continue
		}

		var year int
		if p.format == "date_with_year" {
			year, _ = strconv.Atoi(matches[3])
			// Handle 2-digit years
			if year < 100 {
				year += 2000
			}
		} else {
			// No year specified - use current year or next year if date has passed
			year = refTime.Year()
			tentativeDate := time.Date(year, time.Month(month), day, 0, 0, 0, 0, refTime.Location())
			if tentativeDate.Before(refTime) {
				year++
			}
		}

		// Try to create the date (this validates the date is real)
		date := time.Date(year, time.Month(month), day, 0, 0, 0, 0, refTime.Location())

		// Validate the date is reasonable (not in distant past, within ~2 years future)
		twoYearsFromNow := refTime.AddDate(2, 0, 0)
		if date.After(twoYearsFromNow) {
			continue
		}

		// Remove the date from the title
		cleaned := strings.Replace(text, original, "", 1)
		return &date, original, cleaned
	}

	return nil, "", text
}

// parseDayName handles "monday", "tuesday", "next friday", etc.
func parseDayName(text string, refTime time.Time) (*time.Time, string, string) {
	dayNames := map[string]time.Weekday{
		"sunday":    time.Sunday,
		"monday":    time.Monday,
		"tuesday":   time.Tuesday,
		"wednesday": time.Wednesday,
		"thursday":  time.Thursday,
		"friday":    time.Friday,
		"saturday":  time.Saturday,
		"sun":       time.Sunday,
		"mon":       time.Monday,
		"tue":       time.Tuesday,
		"wed":       time.Wednesday,
		"thu":       time.Thursday,
		"fri":       time.Friday,
		"sat":       time.Saturday,
	}

	// Pattern: "next friday" or "friday"
	pattern := regexp.MustCompile(`(?i)\b(next\s+)?(\w+day)\b`)
	matches := pattern.FindStringSubmatch(text)

	if len(matches) == 0 {
		return nil, "", text
	}

	original := matches[0]
	isNext := strings.TrimSpace(matches[1]) != ""
	dayStr := strings.ToLower(matches[2])

	targetDay, found := dayNames[dayStr]
	if !found {
		return nil, "", text
	}

	// Calculate the target date
	currentDay := refTime.Weekday()
	daysUntil := int(targetDay - currentDay)

	if daysUntil <= 0 || isNext {
		// If the day has passed this week, or "next" is specified, add a week
		daysUntil += 7
		if daysUntil <= 0 {
			daysUntil += 7
		}
	}

	date := refTime.AddDate(0, 0, daysUntil)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, refTime.Location())

	cleaned := strings.Replace(text, original, "", 1)
	return &date, original, cleaned
}

// textToNumber converts text numbers to integers
func textToNumber(text string) (int, bool) {
	numbers := map[string]int{
		"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
		"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
		"eleven": 11, "twelve": 12, "thirteen": 13, "fourteen": 14, "fifteen": 15,
		"sixteen": 16, "seventeen": 17, "eighteen": 18, "nineteen": 19, "twenty": 20,
		"thirty": 30, "forty": 40, "fifty": 50, "sixty": 60, "seventy": 70, "eighty": 80, "ninety": 90,
		"a": 1, "an": 1,
	}

	lower := strings.ToLower(text)
	if num, found := numbers[lower]; found {
		return num, true
	}
	return 0, false
}

// parseRelativeDate handles "today", "tomorrow", "in 3 days", "three days", "2 weeks", etc.
func parseRelativeDate(text string, refTime time.Time) (*time.Time, string, string) {
	// First try text numbers: "in three days", "five days", "a week"
	textPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:in\s+)?(one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|thirteen|fourteen|fifteen|sixteen|seventeen|eighteen|nineteen|twenty|thirty|a|an)\s+(day|days)\b`),
		regexp.MustCompile(`(?i)\b(?:in\s+)?(one|two|three|four|a|an)\s+(week|weeks)\b`),
	}

	for _, pattern := range textPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) >= 3 {
			original := matches[0]
			numText := matches[1]
			unit := strings.ToLower(matches[2])

			if num, ok := textToNumber(numText); ok {
				var daysOffset int
				if strings.HasPrefix(unit, "week") {
					daysOffset = num * 7
				} else {
					daysOffset = num
				}

				date := refTime.AddDate(0, 0, daysOffset)
				date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, refTime.Location())

				cleaned := strings.Replace(text, original, "", 1)
				return &date, original, cleaned
			}
		}
	}

	// Then try numeric relative dates: "in 3 days", "3 days", "2 weeks", "in 1 week", "3 days from now"
	numericPatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:in\s+)?(\d+)\s+(day|days)(?:\s+from\s+now)?\b`),
		regexp.MustCompile(`(?i)\b(?:in\s+)?(\d+)\s+(week|weeks)(?:\s+from\s+now)?\b`),
		regexp.MustCompile(`(?i)\b(?:in\s+)?(\d+)\s+(month|months)(?:\s+from\s+now)?\b`),
	}

	for _, pattern := range numericPatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) >= 3 {
			original := matches[0]
			num, _ := strconv.Atoi(matches[1])
			unit := strings.ToLower(matches[2])

			var daysOffset int
			if strings.HasPrefix(unit, "week") {
				daysOffset = num * 7
			} else if strings.HasPrefix(unit, "month") {
				// Add months using AddDate for correct month arithmetic
				date := refTime.AddDate(0, num, 0)
				date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, refTime.Location())
				cleaned := strings.Replace(text, original, "", 1)
				return &date, original, cleaned
			} else {
				daysOffset = num
			}

			date := refTime.AddDate(0, 0, daysOffset)
			date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, refTime.Location())

			cleaned := strings.Replace(text, original, "", 1)
			return &date, original, cleaned
		}
	}

	// Try time-based relative dates: "in 2 hours", "in 30 minutes", "2 hours from now"
	timePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:in\s+)?(\d+)\s+(hour|hours)(?:\s+from\s+now)?\b`),
		regexp.MustCompile(`(?i)\b(?:in\s+)?(\d+)\s+(minute|minutes|min|mins)(?:\s+from\s+now)?\b`),
	}

	for _, pattern := range timePatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) >= 3 {
			original := matches[0]
			num, _ := strconv.Atoi(matches[1])
			unit := strings.ToLower(matches[2])

			var duration time.Duration
			if strings.HasPrefix(unit, "hour") {
				duration = time.Duration(num) * time.Hour
			} else {
				// minutes
				duration = time.Duration(num) * time.Minute
			}

			date := refTime.Add(duration)
			cleaned := strings.Replace(text, original, "", 1)
			return &date, original, cleaned
		}
	}

	// Try text-based time offsets: "in two hours", "in thirty minutes", "an hour"
	textTimePatterns := []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(?:in\s+)?(one|two|three|four|five|six|seven|eight|nine|ten|eleven|twelve|a|an)\s+(hour|hours)\b`),
		regexp.MustCompile(`(?i)\b(?:in\s+)?(fifteen|thirty|forty-five)\s+(minute|minutes)\b`),
	}

	for _, pattern := range textTimePatterns {
		matches := pattern.FindStringSubmatch(text)
		if len(matches) >= 3 {
			original := matches[0]
			numText := matches[1]
			unit := strings.ToLower(matches[2])

			var num int
			var ok bool
			if numText == "fifteen" {
				num = 15
				ok = true
			} else if numText == "thirty" {
				num = 30
				ok = true
			} else if numText == "forty-five" {
				num = 45
				ok = true
			} else {
				num, ok = textToNumber(numText)
			}

			if ok {
				var duration time.Duration
				if strings.HasPrefix(unit, "hour") {
					duration = time.Duration(num) * time.Hour
				} else {
					// minutes
					duration = time.Duration(num) * time.Minute
				}

				date := refTime.Add(duration)
				cleaned := strings.Replace(text, original, "", 1)
				return &date, original, cleaned
			}
		}
	}

	// Try special period patterns: "end of month", "end of week", "next quarter", "next month", "next year"
	specialPatterns := []struct {
		regex   *regexp.Regexp
		handler func(time.Time) time.Time
	}{
		{
			regex: regexp.MustCompile(`(?i)\bend\s+of\s+(the\s+)?month\b`),
			handler: func(t time.Time) time.Time {
				// Last day of current month
				year, month, _ := t.Date()
				lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, t.Location())
				return lastDay
			},
		},
		{
			regex: regexp.MustCompile(`(?i)\b(start|beginning)\s+of\s+(the\s+)?month\b`),
			handler: func(t time.Time) time.Time {
				// First day of next month
				year, month, _ := t.Date()
				firstDay := time.Date(year, month+1, 1, 0, 0, 0, 0, t.Location())
				return firstDay
			},
		},
		{
			regex: regexp.MustCompile(`(?i)\bmid\s+month\b`),
			handler: func(t time.Time) time.Time {
				// 15th of current month, or next month if passed
				year, month, day := t.Date()
				midMonth := time.Date(year, month, 15, 0, 0, 0, 0, t.Location())
				if day >= 15 {
					// Move to next month
					midMonth = time.Date(year, month+1, 15, 0, 0, 0, 0, t.Location())
				}
				return midMonth
			},
		},
		{
			regex: regexp.MustCompile(`(?i)\bend\s+of\s+(the\s+)?week\b`),
			handler: func(t time.Time) time.Time {
				// Next Sunday (or this Sunday if today is Sunday)
				daysUntilSunday := int((7 - t.Weekday()) % 7)
				if daysUntilSunday == 0 {
					daysUntilSunday = 0 // Today is Sunday
				}
				return time.Date(t.Year(), t.Month(), t.Day()+daysUntilSunday, 0, 0, 0, 0, t.Location())
			},
		},
		{
			regex: regexp.MustCompile(`(?i)\b(start|beginning)\s+of\s+(the\s+)?week\b`),
			handler: func(t time.Time) time.Time {
				// Next Monday
				daysUntilMonday := (8 - int(t.Weekday())) % 7
				if daysUntilMonday == 0 {
					daysUntilMonday = 7
				}
				return time.Date(t.Year(), t.Month(), t.Day()+daysUntilMonday, 0, 0, 0, 0, t.Location())
			},
		},
		{
			regex: regexp.MustCompile(`(?i)\bnext\s+month\b`),
			handler: func(t time.Time) time.Time {
				// Same day next month
				return t.AddDate(0, 1, 0)
			},
		},
		{
			regex: regexp.MustCompile(`(?i)\bnext\s+year\b`),
			handler: func(t time.Time) time.Time {
				// Same day next year
				return t.AddDate(1, 0, 0)
			},
		},
		{
			regex: regexp.MustCompile(`(?i)\bnext\s+quarter\b`),
			handler: func(t time.Time) time.Time {
				// First day of next quarter
				year, month, _ := t.Date()
				currentQuarter := int((month-1)/3) + 1
				nextQuarter := currentQuarter + 1
				nextQuarterYear := year

				if nextQuarter > 4 {
					nextQuarter = 1
					nextQuarterYear++
				}

				nextQuarterMonth := time.Month((nextQuarter-1)*3 + 1)
				return time.Date(nextQuarterYear, nextQuarterMonth, 1, 0, 0, 0, 0, t.Location())
			},
		},
	}

	for _, sp := range specialPatterns {
		if matches := sp.regex.FindString(text); matches != "" {
			date := sp.handler(refTime)
			cleaned := strings.Replace(text, matches, "", 1)
			return &date, matches, cleaned
		}
	}

	// Then try keyword-based relatives
	relatives := map[string]int{
		"today":     0,
		"tomorrow":  1,
		"tmr":       1,
		"yesterday": -1, // Allowed but not recommended
	}

	lowerText := strings.ToLower(text)

	for keyword, daysOffset := range relatives {
		pattern := regexp.MustCompile(`(?i)\b` + keyword + `\b`)
		if pattern.MatchString(lowerText) {
			original := pattern.FindString(text)
			date := refTime.AddDate(0, 0, daysOffset)
			date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, refTime.Location())

			cleaned := strings.Replace(text, original, "", 1)
			return &date, original, cleaned
		}
	}

	return nil, "", text
}

// parseMonthName handles "Nov 26", "November 26", "26 Nov", "26 November"
func parseMonthName(text string, refTime time.Time) (*time.Time, string, string) {
	months := map[string]time.Month{
		"january": time.January, "jan": time.January,
		"february": time.February, "feb": time.February,
		"march": time.March, "mar": time.March,
		"april": time.April, "apr": time.April,
		"may": time.May,
		"june": time.June, "jun": time.June,
		"july": time.July, "jul": time.July,
		"august": time.August, "aug": time.August,
		"september": time.September, "sep": time.September, "sept": time.September,
		"october": time.October, "oct": time.October,
		"november": time.November, "nov": time.November,
		"december": time.December, "dec": time.December,
	}

	// Pattern: "Nov 26" or "26 Nov" or "November 26" or "26 November"
	pattern := regexp.MustCompile(`(?i)\b((\w+)\s+(\d{1,2})|(\d{1,2})\s+(\w+))\b`)
	matches := pattern.FindAllStringSubmatch(text, -1)

	for _, match := range matches {
		if len(match) == 0 {
			continue
		}

		original := match[0]
		var monthStr string
		var dayStr string

		// Check which pattern matched
		if match[2] != "" && match[3] != "" {
			// Month Day format (e.g., "Nov 26")
			monthStr = strings.ToLower(match[2])
			dayStr = match[3]
		} else if match[4] != "" && match[5] != "" {
			// Day Month format (e.g., "26 Nov")
			dayStr = match[4]
			monthStr = strings.ToLower(match[5])
		} else {
			continue
		}

		month, found := months[monthStr]
		if !found {
			continue
		}

		day, err := strconv.Atoi(dayStr)
		if err != nil || day < 1 || day > 31 {
			continue
		}

		// Determine year (current year or next year if date has passed)
		year := refTime.Year()
		tentativeDate := time.Date(year, month, day, 0, 0, 0, 0, refTime.Location())
		if tentativeDate.Before(refTime) {
			year++
		}

		date := time.Date(year, month, day, 0, 0, 0, 0, refTime.Location())

		// Validate the date is reasonable
		twoYearsFromNow := refTime.AddDate(2, 0, 0)
		if date.After(twoYearsFromNow) {
			continue
		}

		cleaned := strings.Replace(text, original, "", 1)
		return &date, original, cleaned
	}

	return nil, "", text
}

// parseTime handles time expressions like "at 3pm", "3:30pm", "15:00", "at 3:30"
func parseTime(text string) (*time.Time, string, string) {
	// Time patterns to try
	patterns := []struct {
		regex   *regexp.Regexp
		handler func([]string, time.Time) (*time.Time, bool)
	}{
		{
			// "at 3pm", "at 3:30pm", "at 3:30 pm"
			regex: regexp.MustCompile(`(?i)\bat\s+(\d{1,2})(?::(\d{2}))?\s?(am|pm)\b`),
			handler: func(matches []string, ref time.Time) (*time.Time, bool) {
				hour, _ := strconv.Atoi(matches[1])
				min := 0
				if matches[2] != "" {
					min, _ = strconv.Atoi(matches[2])
				}
				ampm := strings.ToLower(matches[3])

				// Convert to 24-hour format
				if ampm == "pm" && hour != 12 {
					hour += 12
				} else if ampm == "am" && hour == 12 {
					hour = 0
				}

				if hour < 0 || hour > 23 || min < 0 || min > 59 {
					return nil, false
				}

				t := time.Date(ref.Year(), ref.Month(), ref.Day(), hour, min, 0, 0, ref.Location())
				return &t, true
			},
		},
		{
			// "3pm", "3:30pm", "3:30 pm" (without "at")
			regex: regexp.MustCompile(`(?i)\b(\d{1,2})(?::(\d{2}))?\s?(am|pm)\b`),
			handler: func(matches []string, ref time.Time) (*time.Time, bool) {
				hour, _ := strconv.Atoi(matches[1])
				min := 0
				if matches[2] != "" {
					min, _ = strconv.Atoi(matches[2])
				}
				ampm := strings.ToLower(matches[3])

				// Convert to 24-hour format
				if ampm == "pm" && hour != 12 {
					hour += 12
				} else if ampm == "am" && hour == 12 {
					hour = 0
				}

				if hour < 0 || hour > 23 || min < 0 || min > 59 {
					return nil, false
				}

				t := time.Date(ref.Year(), ref.Month(), ref.Day(), hour, min, 0, 0, ref.Location())
				return &t, true
			},
		},
		{
			// "at 15:00", "at 3:30" (24-hour format or without am/pm)
			regex: regexp.MustCompile(`(?i)\bat\s+(\d{1,2}):(\d{2})\b`),
			handler: func(matches []string, ref time.Time) (*time.Time, bool) {
				hour, _ := strconv.Atoi(matches[1])
				min, _ := strconv.Atoi(matches[2])

				if hour < 0 || hour > 23 || min < 0 || min > 59 {
					return nil, false
				}

				t := time.Date(ref.Year(), ref.Month(), ref.Day(), hour, min, 0, 0, ref.Location())
				return &t, true
			},
		},
	}

	now := time.Now()
	for _, p := range patterns {
		matches := p.regex.FindStringSubmatch(text)
		if len(matches) > 0 {
			original := matches[0]
			if timeVal, ok := p.handler(matches, now); ok {
				cleaned := strings.Replace(text, original, "", 1)
				return timeVal, original, cleaned
			}
		}
	}

	return nil, "", text
}
