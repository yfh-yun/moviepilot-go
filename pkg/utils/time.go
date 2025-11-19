package utils

import (
	"fmt"
	"strings"
	"time"
)

// Now returns the current time in UTC
func Now() time.Time {
	return time.Now().UTC()
}

// NowTimestamp returns the current Unix timestamp
func NowTimestamp() int64 {
	return time.Now().Unix()
}

// NowMilliseconds returns the current Unix timestamp in milliseconds
func NowMilliseconds() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// ParseTime parses a time string using the specified format
func ParseTime(timeStr, format string) (time.Time, error) {
	return time.Parse(format, timeStr)
}

// FormatTime formats a time.Time using the specified format
func FormatTime(t time.Time, format string) string {
	return t.Format(format)
}

// FormatTimestamp formats a Unix timestamp using the specified format
func FormatTimestamp(timestamp int64, format string) string {
	t := time.Unix(timestamp, 0)
	return t.Format(format)
}

// FormatDuration formats a duration in a human-readable way
func FormatDuration(duration time.Duration) string {
	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	}

	if duration < time.Hour {
		minutes := int(duration.Minutes())
		seconds := int(duration.Seconds()) % 60
		return fmt.Sprintf("%dm %ds", minutes, seconds)
	}

	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}

// DaysAgo returns the time that was n days ago
func DaysAgo(n int) time.Time {
	return time.Now().AddDate(0, 0, -n)
}

// DaysFromNow returns the time that will be n days from now
func DaysFromNow(n int) time.Time {
	return time.Now().AddDate(0, 0, n)
}

// HoursAgo returns the time that was n hours ago
func HoursAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * time.Hour)
}

// HoursFromNow returns the time that will be n hours from now
func HoursFromNow(n int) time.Time {
	return time.Now().Add(time.Duration(n) * time.Hour)
}

// MinutesAgo returns the time that was n minutes ago
func MinutesAgo(n int) time.Time {
	return time.Now().Add(-time.Duration(n) * time.Minute)
}

// MinutesFromNow returns the time that will be n minutes from now
func MinutesFromNow(n int) time.Time {
	return time.Now().Add(time.Duration(n) * time.Minute)
}

// IsToday checks if a time is today
func IsToday(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}

// IsThisWeek checks if a time is this week
func IsThisWeek(t time.Time) bool {
	now := time.Now()
	year1, week1 := now.ISOWeek()
	year2, week2 := t.ISOWeek()
	return year1 == year2 && week1 == week2
}

// IsThisMonth checks if a time is this month
func IsThisMonth(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month()
}

// IsThisYear checks if a time is this year
func IsThisYear(t time.Time) bool {
	return t.Year() == time.Now().Year()
}

// StartOfDay returns the start of the day for a given time
func StartOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

// EndOfDay returns the end of the day for a given time
func EndOfDay(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 23, 59, 59, 999999999, t.Location())
}

// StartOfWeek returns the start of the week (Monday) for a given time
func StartOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	monday := t.AddDate(0, 0, -int(weekday)+1)
	return StartOfDay(monday)
}

// EndOfWeek returns the end of the week (Sunday) for a given time
func EndOfWeek(t time.Time) time.Time {
	weekday := t.Weekday()
	if weekday == time.Sunday {
		weekday = 7
	}
	sunday := t.AddDate(0, 0, 7-int(weekday))
	return EndOfDay(sunday)
}

// StartOfMonth returns the start of the month for a given time
func StartOfMonth(t time.Time) time.Time {
	year, month, _ := t.Date()
	return time.Date(year, month, 1, 0, 0, 0, 0, t.Location())
}

// EndOfMonth returns the end of the month for a given time
func EndOfMonth(t time.Time) time.Time {
	return StartOfMonth(t).AddDate(0, 1, -1)
}

// HumanizeTime returns a human-readable relative time string
func HumanizeTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	if diff < time.Minute {
		return "just now"
	}

	if diff < time.Hour {
		minutes := int(diff.Minutes())
		if minutes == 1 {
			return "1 minute ago"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	}

	if diff < 24*time.Hour {
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", hours)
	}

	if diff < 7*24*time.Hour {
		days := int(diff.Hours() / 24)
		if days == 1 {
			return "1 day ago"
		}
		return fmt.Sprintf("%d days ago", days)
	}

	if diff < 30*24*time.Hour {
		weeks := int(diff.Hours() / (7 * 24))
		if weeks == 1 {
			return "1 week ago"
		}
		return fmt.Sprintf("%d weeks ago", weeks)
	}

	if diff < 365*24*time.Hour {
		months := int(diff.Hours() / (30 * 24))
		if months == 1 {
			return "1 month ago"
		}
		return fmt.Sprintf("%d months ago", months)
	}

	years := int(diff.Hours() / (365 * 24))
	if years == 1 {
		return "1 year ago"
	}
	return fmt.Sprintf("%d years ago", years)
}

// IsLeapYear checks if a year is a leap year
func IsLeapYear(year int) bool {
	return year%4 == 0 && (year%100 != 0 || year%400 == 0)
}

// DaysInMonth returns the number of days in a month for a given year
func DaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

// WeekNumber returns the ISO week number for a given time
func WeekNumber(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// IsWeekend checks if a time is on a weekend
func IsWeekend(t time.Time) bool {
	weekday := t.Weekday()
	return weekday == time.Saturday || weekday == time.Sunday
}

// IsWeekday checks if a time is on a weekday
func IsWeekday(t time.Time) bool {
	return !IsWeekend(t)
}

// TimeZoneList returns a list of common timezones
func TimeZoneList() []string {
	return []string{
		"UTC",
		"America/New_York",
		"America/Los_Angeles",
		"America/Chicago",
		"Europe/London",
		"Europe/Paris",
		"Asia/Tokyo",
		"Asia/Shanghai",
		"Australia/Sydney",
	}
}

// SetTimeZone sets the timezone for a time
func SetTimeZone(t time.Time, timezone string) (time.Time, error) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return t, err
	}
	return t.In(loc), nil
}

// TimeDiff returns the difference between two times in various units
func TimeDiff(t1, t2 time.Time) (days, hours, minutes, seconds int) {
	diff := t2.Sub(t1)

	days = int(diff.Hours() / 24)
	hours = int(diff.Hours()) % 24
	minutes = int(diff.Minutes()) % 60
	seconds = int(diff.Seconds()) % 60

	return days, hours, minutes, seconds
}

// ParseDuration parses a duration string (e.g., "1h30m")
func ParseDuration(durationStr string) (time.Duration, error) {
	return time.ParseDuration(durationStr)
}

// FormatDurationString formats a duration as a string (e.g., "1h30m")
func FormatDurationString(duration time.Duration) string {
	if duration < time.Second {
		return duration.String()
	}

	parts := []string{}

	// Hours
	hours := int(duration.Hours())
	if hours > 0 {
		parts = append(parts, fmt.Sprintf("%dh", hours))
		duration -= time.Duration(hours) * time.Hour
	}

	// Minutes
	minutes := int(duration.Minutes())
	if minutes > 0 {
		parts = append(parts, fmt.Sprintf("%dm", minutes))
		duration -= time.Duration(minutes) * time.Minute
	}

	// Seconds
	seconds := int(duration.Seconds())
	if seconds > 0 {
		parts = append(parts, fmt.Sprintf("%ds", seconds))
	}

	return strings.Join(parts, "")
}

// SleepWithContext sleeps for the specified duration, but can be canceled by context
func SleepWithContext(duration time.Duration) error {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	// This is a simplified version - actual context-aware sleep would use context
	<-timer.C
	return nil
}

// TimestampToTime converts a Unix timestamp to time.Time
func TimestampToTime(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

// TimeToTimestamp converts a time.Time to Unix timestamp
func TimeToTimestamp(t time.Time) int64 {
	return t.Unix()
}

// TimeToMilliseconds converts a time.Time to milliseconds since Unix epoch
func TimeToMilliseconds(t time.Time) int64 {
	return t.UnixNano() / int64(time.Millisecond)
}

// MillisecondsToTime converts milliseconds since Unix epoch to time.Time
func MillisecondsToTime(milliseconds int64) time.Time {
	return time.Unix(0, milliseconds*int64(time.Millisecond))
}

// Age calculates the age based on birth date
func Age(birthDate time.Time) int {
	now := time.Now()
	age := now.Year() - birthDate.Year()

	if now.Month() < birthDate.Month() ||
		(now.Month() == birthDate.Month() && now.Day() < birthDate.Day()) {
		age--
	}

	return age
}
