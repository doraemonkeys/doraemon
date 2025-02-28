package doraemon

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// HexToInt converts a hexadecimal string to an int64 value.
// If an error occurs during conversion, the function panics.
func HexToInt(hex string) int64 {
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = hex[2:]
	}
	num, err := strconv.ParseInt(hex, 16, 64)
	if err != nil {
		panic(err)
	}
	return num
}

func HexToBigInt(hex string) *big.Int {
	if strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X") {
		hex = hex[2:]
	}
	num := new(big.Int)
	_, ok := num.SetString(hex, 16)
	if !ok {
		panic("invalid hex string")
	}
	return num
}

func HexToInt2(hex string) (int64, error) {
	if len(hex) > 2 && (strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X")) {
		hex = hex[2:]
	}
	var result int64
	for _, v := range hex {
		result *= 16
		switch {
		case v >= '0' && v <= '9':
			result += int64(v - '0')
		case v >= 'a' && v <= 'f':
			result += int64(v - 'a' + 10)
		case v >= 'A' && v <= 'F':
			result += int64(v - 'A' + 10)
		default:
			return 0, errors.New("invalid hex string")
		}
	}
	return result, nil
}

// CheckInRecorder is a daily check-in recorder that resets at midnight.
// It's not thread-safe and requires external locking for concurrent access.
// It uses a uint64 to store the check-in status, each bit represents a day's check-in status: 0 for not checked in, 1 for checked in.
// It can record up to 64 days.
type CheckInRecorder struct {
	// record stores the check-in status. Each bit represents a day's check-in status: 0 for not checked in, 1 for checked in.
	// It can record up to 64 days.
	record          uint64
	lastCheckInTime time.Time
	location        *time.Location
	clock           func() time.Time
}

// NewEmptyCheckInRecorder creates a new CheckInRecorder with no check-in history.
func NewEmptyCheckInRecorder(location *time.Location) *CheckInRecorder {
	return &CheckInRecorder{
		record:          0,
		location:        location,
		lastCheckInTime: time.Time{}.In(location),
		clock:           time.Now,
	}
}

// Integer is a type constraint for integer types.
type Integer interface {
	uint64 | uint32 | uint16 | uint8 | int64 | int32 | int16 | int8
}

// newCheckInRecorder creates a new CheckInRecorder with existing record and last check-in time.
func NewCheckInRecorder[T Integer](recordRaw T, lastCheckInTime time.Time) *CheckInRecorder {
	recorder := newCheckInRecorder(recordRaw, lastCheckInTime)
	diffDays := calculateDaysDiff(recorder.lastCheckInTime, recorder.clock().In(recorder.location))
	if diffDays <= 0 && recorder.record&1 == 0 {
		panic("already checked in today, but today's record is 0")
	}
	return recorder
}

func newCheckInRecorder[T Integer](recordRaw T, lastCheckInTime time.Time) *CheckInRecorder {
	if lastCheckInTime.IsZero() && recordRaw != 0 {
		panic("lastCheckInTime is zero, but recordRaw is not zero")
	}
	// var recorder *CheckInRecorder
	if recordRaw >= 0 {
		return &CheckInRecorder{
			record:          uint64(recordRaw),
			lastCheckInTime: lastCheckInTime,
			location:        lastCheckInTime.Location(),
			clock:           time.Now,
		}
	}
	var record uint64
	switch any(recordRaw).(type) {
	case int64:
		record = uint64(recordRaw)
	case int32:
		record = uint64(uint32(recordRaw))
	case int16:
		record = uint64(uint16(recordRaw))
	case int8:
		record = uint64(uint8(recordRaw))
	default:
		panic("unreachable")
	}
	return &CheckInRecorder{
		record:          record,
		lastCheckInTime: lastCheckInTime,
		location:        lastCheckInTime.Location(),
		clock:           time.Now,
	}
}

func (s *CheckInRecorder) SetClock(clock func() time.Time) {
	s.clock = clock
}

// CheckIn records a new check-in.
// It returns false if the user has already checked in today.
func (s *CheckInRecorder) CheckIn() bool {
	now := s.clock().In(s.location)
	if !s.correctCheckInRecord(now) {
		s.record |= 1
		s.lastCheckInTime = now
		return true
	}
	return false
}

// correctCheckInRecord corrects the check-in record based on the current time.
// It returns true if the user has already checked in today, false otherwise.
func (s *CheckInRecorder) correctCheckInRecord(now time.Time) (checkedIn bool) {
	if s.record == 0 {
		// Already cleared, no need to reset.
		return false
	}
	// Calculate the difference in days between the last check-in time and today's midnight.
	diffDays := calculateDaysDiff(s.lastCheckInTime, now)
	if diffDays <= 0 {
		// Already checked in today.
		return true
	}
	if diffDays > 64 {
		s.record = 0
		return false
	}
	s.record <<= uint64(diffDays)
	return false
}

// calculateDaysDiff calculates the difference in days between the last check-in time and today's midnight, considering daylight saving time (DST).
// The caller must ensure that lastCheckIn and now are in the same timezone.
func calculateDaysDiff(lastCheckIn time.Time, now time.Time) int {
	// Get today's midnight in the same location
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, lastCheckIn.Location())

	// Calculate the difference between today's midnight and the last check-in time
	diff := todayMidnight.Sub(lastCheckIn)
	lastCheckInIsDST := lastCheckIn.IsDST()
	nowIsDST := now.IsDST()
	if !lastCheckInIsDST && nowIsDST {
		diff += time.Hour
		return int(diff.Hours()/24) + 1
	} else if lastCheckInIsDST && !nowIsDST {
		diff -= time.Hour
		return int(diff.Hours()/24) + 1
	}
	if diff <= 0 {
		return 0
	}
	return int(diff.Hours()/24) + 1
}

// RawRecord returns the raw record value and last check-in time.
func (s *CheckInRecorder) RawRecord() (record uint64, lastCheckInTime time.Time) {
	s.correctCheckInRecord(s.clock().In(s.location))
	return s.record, s.lastCheckInTime
}

// Record returns the check-in record as a boolean slice.
func (s *CheckInRecorder) Record() []bool {
	return s.RecordN(64)
}

// RecordN returns the check-in record as a boolean slice with the specified length.
func (s *CheckInRecorder) RecordN(n int) []bool {
	if n < 0 || n > 64 {
		panic("n must be between 0 and 64")
	}
	s.correctCheckInRecord(s.clock().In(s.location))
	record := make([]bool, n)
	for i := range n {
		record[i] = (s.record & (1 << i)) != 0
	}
	return record
}

// HasCheckIn checks if the user has checked in on the specified day (0-based index).
func (s *CheckInRecorder) HasCheckIn(day int) bool {
	if day < 0 || day >= 64 {
		return false
	}
	s.correctCheckInRecord(s.clock().In(s.location))
	return (s.record & (1 << day)) != 0
}

// HasCheckedInToday checks if the user has already checked in today.
func (s *CheckInRecorder) HasCheckedInToday() bool {
	return s.correctCheckInRecord(s.clock().In(s.location))
}

// ConsecutiveCheckInDays returns the number of consecutive check-in days.
func (s *CheckInRecorder) ConsecutiveCheckInDays() int {
	s.correctCheckInRecord(s.clock().In(s.location))
	consecutiveDays := 0
	for i := range 64 {
		if s.record&(1<<i) == 0 {
			break
		}
		consecutiveDays++
	}
	return consecutiveDays
}

// DaysToNextMilestone returns the remaining days to reach the next milestone (e.g., 7-day, 30-day).
func (s *CheckInRecorder) DaysToNextMilestone(milestone int) int {
	consecutive := s.ConsecutiveCheckInDays()
	if consecutive >= milestone {
		return 0
	}
	return milestone - consecutive
}

func newCheckInRecorderForTest(record any, lastCheckInTime time.Time) *CheckInRecorder {
	switch any(record).(type) {
	case uint64:
		return NewCheckInRecorder(record.(uint64), lastCheckInTime)
	case uint32:
		return NewCheckInRecorder(record.(uint32), lastCheckInTime)
	case uint16:
		return NewCheckInRecorder(record.(uint16), lastCheckInTime)
	case uint8:
		return NewCheckInRecorder(record.(uint8), lastCheckInTime)
	case int64:
		return NewCheckInRecorder(record.(int64), lastCheckInTime)
	case int32:
		return NewCheckInRecorder(record.(int32), lastCheckInTime)
	case int16:
		return NewCheckInRecorder(record.(int16), lastCheckInTime)
	case int8:
		return NewCheckInRecorder(record.(int8), lastCheckInTime)
	default:
		panic("unreachable")
	}
}
