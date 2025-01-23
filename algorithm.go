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

// SignInRecorder is a daily sign-in recorder that resets at midnight.
// It's not thread-safe and requires external locking for concurrent access.
// It uses a uint64 to store the sign-in status, each bit represents a day's sign-in status: 0 for not signed in, 1 for signed in.
// It can record up to 64 days.
type SignInRecorder struct {
	// record stores the sign-in status. Each bit represents a day's sign-in status: 0 for not signed in, 1 for signed in.
	// It can record up to 64 days.
	record         uint64
	lastSignInTime time.Time
	location       *time.Location
	clock          func() time.Time
}

// NewEmptySignInRecorder creates a new SignInRecorder with no sign-in history.
func NewEmptySignInRecorder(location *time.Location) *SignInRecorder {
	return &SignInRecorder{
		record:         0,
		location:       location,
		lastSignInTime: time.Time{}.In(location),
		clock:          time.Now,
	}
}

// Integer is a type constraint for integer types.
type Integer interface {
	~uint64 | ~uint32 | ~uint16 | ~uint8 | ~int64 | ~int32 | ~int16 | ~int8
}

// NewSignInRecorder creates a new SignInRecorder with existing record and last sign-in time.
func NewSignInRecorder[T Integer](recordRaw T, lastSignInTime time.Time) *SignInRecorder {
	if recordRaw >= 0 {
		return &SignInRecorder{
			record:         uint64(recordRaw),
			lastSignInTime: lastSignInTime,
			location:       lastSignInTime.Location(),
			clock:          time.Now,
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
	return &SignInRecorder{
		record:         record,
		lastSignInTime: lastSignInTime,
		location:       lastSignInTime.Location(),
		clock:          time.Now,
	}
}

func (s *SignInRecorder) SetClock(clock func() time.Time) {
	s.clock = clock
}

// SignIn records a new sign-in.
// It returns false if the user has already signed in today.
func (s *SignInRecorder) SignIn() bool {
	now := s.clock().In(s.location)
	if !s.correctSignInRecord(now) {
		s.record |= 1
		s.lastSignInTime = now
		return true
	}
	return false
}

// correctSignInRecord corrects the sign-in record based on the current time.
// It returns true if the user has already signed in today, false otherwise.
func (s *SignInRecorder) correctSignInRecord(now time.Time) (signed bool) {
	if s.record == 0 {
		// Already cleared, no need to reset.
		return false
	}
	// Calculate the difference in days between the last sign-in time and today's midnight.
	diffDays := calculateDaysDiff(s.lastSignInTime, now)
	if diffDays <= 0 {
		// Already signed in today.
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

// RawRecord returns the raw record value and last sign-in time.
func (s *SignInRecorder) RawRecord() (record uint64, lastSignInTime time.Time) {
	s.correctSignInRecord(s.clock().In(s.location))
	return s.record, s.lastSignInTime
}

// Record returns the sign-in record as a boolean slice.
func (s *SignInRecorder) Record() []bool {
	return s.RecordN(64)
}

// RecordN returns the sign-in record as a boolean slice with the specified length.
func (s *SignInRecorder) RecordN(n int) []bool {
	if n < 0 || n > 64 {
		panic("n must be between 0 and 64")
	}
	s.correctSignInRecord(s.clock().In(s.location))
	record := make([]bool, n)
	for i := 0; i < n; i++ {
		record[i] = (s.record & (1 << i)) != 0
	}
	return record
}

// HasSignIn checks if the user has signed in on the specified day (0-based index).
func (s *SignInRecorder) HasSignIn(day int) bool {
	if day < 0 || day >= 64 {
		return false
	}
	s.correctSignInRecord(s.clock().In(s.location))
	return (s.record & (1 << day)) != 0
}

// HasSignedToday checks if the user has already signed in today.
func (s *SignInRecorder) HasSignedToday() bool {
	return s.correctSignInRecord(s.clock().In(s.location))
}

// ConsecutiveSignInDays returns the number of consecutive sign-in days.
func (s *SignInRecorder) ConsecutiveSignInDays() int {
	s.correctSignInRecord(s.clock().In(s.location))
	consecutiveDays := 0
	for i := 0; i < 64; i++ {
		if s.record&(1<<i) == 0 {
			break
		}
		consecutiveDays++
	}
	return consecutiveDays
}

// DaysToNextMilestone returns the remaining days to reach the next milestone (e.g., 7-day, 30-day).
func (s *SignInRecorder) DaysToNextMilestone(milestone int) int {
	consecutive := s.ConsecutiveSignInDays()
	if consecutive >= milestone {
		return 0
	}
	return milestone - consecutive
}

func newSignInRecorderForTest(record any, lastSignInTime time.Time) *SignInRecorder {
	switch any(record).(type) {
	case uint64:
		return NewSignInRecorder(record.(uint64), lastSignInTime)
	case uint32:
		return NewSignInRecorder(record.(uint32), lastSignInTime)
	case uint16:
		return NewSignInRecorder(record.(uint16), lastSignInTime)
	case uint8:
		return NewSignInRecorder(record.(uint8), lastSignInTime)
	case int64:
		return NewSignInRecorder(record.(int64), lastSignInTime)
	case int32:
		return NewSignInRecorder(record.(int32), lastSignInTime)
	case int16:
		return NewSignInRecorder(record.(int16), lastSignInTime)
	case int8:
		return NewSignInRecorder(record.(int8), lastSignInTime)
	default:
		panic("unreachable")
	}
}
