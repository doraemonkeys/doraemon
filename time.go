package doraemon

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// FormatUnixMilliString parses a Unix millisecond timestamp from a string and formats it.
func FormatUnixMilliString(unixMilliStr string, format string) (string, error) {
	if unixMilliStr == "" {
		return "", fmt.Errorf("unix milli string is empty")
	}

	unixMilli, err := strconv.ParseInt(unixMilliStr, 10, 64)
	if err != nil {
		return "", err
	}

	formattedTime := time.UnixMilli(unixMilli).In(time.Local).Format(format)
	return formattedTime, nil
}

// FormatUnixMilli converts a Unix millisecond timestamp (int64) into a formatted string
// in the local timezone.
func FormatUnixMilli(unixMilli int64, format string) string {
	return time.UnixMilli(unixMilli).In(time.Local).Format(format)
}

// TrackTime calculates the elapsed time since the given time and prints it.
//
// Usage:
//
//	defer TrackTime(time.Now())
func TrackTime(pre time.Time, message ...string) time.Duration {
	if len(message) == 0 {
		message = append(message, "Time elapsed:")
	}
	elapsed := time.Since(pre)
	out := strings.Join(message, " ") + " " + elapsed.String()
	fmt.Println(out)
	return elapsed
}
