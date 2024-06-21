package doraemon

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// 时间戳毫秒 -> format time(Local)。
// if rawTime is "" , return current time。
func GetFormatedTimeFromUnixMilliStr(rawTime string, format string) (string, error) {
	if rawTime == "" || rawTime == "0" {
		return time.Now().Format(format), nil
	}
	// 时间戳秒 -> format time
	latestTime, err := strconv.ParseInt(rawTime, 10, 64)
	if err != nil {
		return "", err
	}
	foramtedTime := time.UnixMilli(int64(latestTime)).In(time.Local).Format(format)
	return foramtedTime, nil
}

// 时间戳毫秒 -> format time(Local)。
// if rawTime is "" , return current time。
func GetFormatedTimeFromUnixMilli(rawTime int64, format string) string {
	return time.UnixMilli(rawTime).In(time.Local).Format(format)
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
