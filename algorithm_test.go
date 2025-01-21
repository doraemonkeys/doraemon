package doraemon

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestHexToInt2(t *testing.T) {
	tests := []struct {
		hex      string
		expected int64
		err      error
	}{
		{hex: "0", expected: 0, err: nil},
		{hex: "1", expected: 1, err: nil},
		{hex: "A", expected: 10, err: nil},
		{hex: "F", expected: 15, err: nil},
		{hex: "10", expected: 16, err: nil},
		{hex: "FF", expected: 255, err: nil},
		{hex: "100", expected: 256, err: nil},
		{hex: "ABC", expected: 2748, err: nil},
		{hex: "0x10", expected: 16, err: nil},
		{hex: "0xFF", expected: 255, err: nil},
		{hex: "0xABC", expected: 2748, err: nil},
		{hex: "0xaBC", expected: 2748, err: nil},
		{hex: "0x", expected: 0, err: errors.New("invalid hex string")},
		{hex: "G", expected: 0, err: errors.New("invalid hex string")},
	}

	for _, tt := range tests {
		actual, err := HexToInt2(tt.hex)
		if actual != tt.expected {
			t.Errorf("HexToInt2(%s) = %d, expected %d", tt.hex, actual, tt.expected)
		}
		if (err != nil && tt.err == nil) || (err == nil && tt.err != nil) || (err != nil && tt.err != nil && err.Error() != tt.err.Error()) {
			t.Errorf("HexToInt2(%s) error = %v, expected %v", tt.hex, err, tt.err)
		}
	}
}

func TestHexToInt(t *testing.T) {
	tests := []struct {
		hex      string
		expected int64
		err      error
	}{
		{hex: "0", expected: 0, err: nil},
		{hex: "1", expected: 1, err: nil},
		{hex: "A", expected: 10, err: nil},
		{hex: "F", expected: 15, err: nil},
		{hex: "10", expected: 16, err: nil},
		{hex: "FF", expected: 255, err: nil},
		{hex: "100", expected: 256, err: nil},
		{hex: "ABC", expected: 2748, err: nil},
		{hex: "0x10", expected: 16, err: nil},
		{hex: "0xFF", expected: 255, err: nil},
		{hex: "0xABC", expected: 2748, err: nil},
		{hex: "0xaBC", expected: 2748, err: nil},
		{hex: "0x", expected: 0, err: errors.New("invalid hex string")},
		{hex: "G", expected: 0, err: errors.New("invalid hex string")},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					if tt.err == nil {
						t.Errorf("HexToInt(%s) panic: %v, expected no panic", tt.hex, r)
					}
				}
			}()
			actual := HexToInt(tt.hex)
			if actual != tt.expected {
				t.Errorf("HexToInt(%s) = %d, expected %d", tt.hex, actual, tt.expected)
			}
		})
	}
}

func TestHexToBigInt(t *testing.T) {
	tests := []struct {
		hex      string
		expected string
	}{
		{"0x1", "1"},
		{"0X1", "1"},
		{"1", "1"},
		{"0xA", "10"},
		{"0x10", "16"},
		{"123ABC", "1194684"},
		{"0x123ABC", "1194684"},
		{"abcdef", "11259375"},
		{"0xabcdef", "11259375"},
	}

	for _, tt := range tests {
		t.Run(tt.hex, func(t *testing.T) {
			got := HexToBigInt(tt.hex).String()
			if got != tt.expected {
				t.Errorf("HexToBigInt(%s) = %s; want %s", tt.hex, got, tt.expected)
			}
		})
	}
}

// This will test panic for invalid input strings.
func TestHexToBigIntInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for invalid hex string, but did not get one")
		}
	}()
	HexToBigInt("invalid")
}

func Test_calculateDaysDiff(t *testing.T) {
	new_york, _ := time.LoadLocation("America/New_York")
	shanghai, _ := time.LoadLocation("Asia/Shanghai")
	tests := []struct {
		name        string
		lastCheckIn time.Time
		now         time.Time
		want        int
	}{
		{name: "1", lastCheckIn: time.Date(2025, 1, 1, 11, 23, 0, 0, time.UTC), now: time.Date(2025, 1, 1, 12, 8, 0, 0, time.UTC), want: 0},
		{name: "2", lastCheckIn: time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC), now: time.Date(2025, 1, 2, 12, 0, 0, 0, time.UTC), want: 1},
		{name: "3", lastCheckIn: time.Date(2025, 7, 8, 23, 59, 0, 0, time.UTC), now: time.Date(2025, 7, 9, 20, 59, 0, 0, time.UTC), want: 1},
		{name: "4", lastCheckIn: time.Date(2025, 7, 8, 23, 59, 0, 0, time.UTC), now: time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC), want: 2},
		{name: "5", lastCheckIn: time.Date(2025, 7, 8, 23, 59, 0, 0, time.UTC), now: time.Date(2025, 7, 10, 12, 0, 0, 0, time.UTC), want: 2},
		{name: "6", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 11, 0, 0, 0, 0, new_york), want: 2},
		{name: "7", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 11, 1, 0, 0, 0, new_york), want: 2},
		{name: "8", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 11, 23, 59, 0, 0, new_york), want: 2},
		{name: "9", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 12, 23, 59, 0, 0, new_york), want: 3},
		{name: "10", lastCheckIn: time.Date(2024, 3, 8, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 12, 23, 59, 0, 0, new_york), want: 4},
		{name: "11", lastCheckIn: time.Date(2024, 3, 8, 23, 58, 0, 0, time.UTC), now: time.Date(2024, 3, 12, 23, 59, 0, 0, time.UTC), want: 4},
		{name: "8", lastCheckIn: time.Date(2024, 3, 19, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 21, 23, 59, 0, 0, new_york), want: 2},
		{name: "9", lastCheckIn: time.Date(2024, 3, 19, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 22, 23, 59, 0, 0, new_york), want: 3},
		{name: "10", lastCheckIn: time.Date(2024, 3, 18, 23, 58, 0, 0, new_york), now: time.Date(2024, 3, 22, 23, 59, 0, 0, new_york), want: 4},
		{name: "11", lastCheckIn: time.Date(2024, 3, 18, 23, 58, 0, 0, time.UTC), now: time.Date(2024, 3, 22, 23, 59, 0, 0, time.UTC), want: 4},

		{name: "8", lastCheckIn: time.Date(2024, 11, 2, 23, 58, 0, 0, new_york), now: time.Date(2024, 11, 3, 23, 59, 0, 0, new_york), want: 1},
		{name: "9", lastCheckIn: time.Date(2024, 11, 2, 23, 58, 0, 0, new_york), now: time.Date(2024, 11, 4, 23, 59, 0, 0, new_york), want: 2},
		{name: "10", lastCheckIn: time.Date(2024, 11, 1, 23, 58, 0, 0, new_york), now: time.Date(2024, 11, 4, 23, 59, 0, 0, new_york), want: 3},
		{name: "11", lastCheckIn: time.Date(2024, 11, 1, 23, 58, 0, 0, time.UTC), now: time.Date(2024, 11, 4, 23, 59, 0, 0, time.UTC), want: 3},

		{name: "8", lastCheckIn: time.Date(2024, 11, 2, 0, 2, 0, 0, new_york), now: time.Date(2024, 11, 4, 23, 59, 0, 0, new_york), want: 2},

		{name: "dasd", lastCheckIn: time.Date(2025, 1, 1, 11, 23, 0, 0, shanghai), now: time.Date(2025, 1, 1, 12, 8, 0, 0, shanghai), want: 0},
		{name: "d34as", lastCheckIn: time.Date(2025, 1, 1, 12, 0, 0, 0, shanghai), now: time.Date(2025, 1, 2, 12, 0, 0, 0, shanghai), want: 1},
		{name: "3v345", lastCheckIn: time.Date(2025, 7, 8, 23, 59, 0, 0, shanghai), now: time.Date(2025, 7, 9, 20, 59, 0, 0, shanghai), want: 1},
		{name: "4xvqaw", lastCheckIn: time.Date(2025, 7, 8, 23, 59, 0, 0, shanghai), now: time.Date(2025, 7, 10, 0, 0, 0, 0, shanghai), want: 2},
		{name: "5g4t4", lastCheckIn: time.Date(2025, 7, 8, 23, 59, 0, 0, shanghai), now: time.Date(2025, 7, 10, 12, 0, 0, 0, shanghai), want: 2},
		{name: "6gd343t4", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, shanghai), now: time.Date(2024, 3, 11, 0, 0, 0, 0, shanghai), want: 2},
		{name: "7fdsvx", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, shanghai), now: time.Date(2024, 3, 11, 1, 0, 0, 0, shanghai), want: 2},
		{name: "8bvcbcv", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, shanghai), now: time.Date(2024, 3, 11, 23, 59, 0, 0, shanghai), want: 2},
		{name: "9fg32r", lastCheckIn: time.Date(2024, 3, 9, 23, 58, 0, 0, shanghai), now: time.Date(2024, 3, 12, 23, 59, 0, 0, shanghai), want: 3},
		{name: "1fuyu0", lastCheckIn: time.Date(2024, 3, 8, 23, 58, 0, 0, shanghai), now: time.Date(2024, 3, 12, 23, 59, 0, 0, shanghai), want: 4},
		{name: "11hfg476", lastCheckIn: time.Date(2024, 3, 8, 23, 58, 0, 0, shanghai), now: time.Date(2024, 3, 12, 23, 59, 0, 0, shanghai), want: 4},
		{name: "8eqwr45", lastCheckIn: time.Date(2024, 11, 2, 23, 58, 0, 0, shanghai), now: time.Date(2024, 11, 3, 23, 59, 0, 0, shanghai), want: 1},
		{name: "9656yhdtxg", lastCheckIn: time.Date(2024, 11, 2, 23, 58, 0, 0, shanghai), now: time.Date(2024, 11, 4, 23, 59, 0, 0, shanghai), want: 2},
		{name: "1wdad0", lastCheckIn: time.Date(2024, 11, 1, 23, 58, 0, 0, shanghai), now: time.Date(2024, 11, 4, 23, 59, 0, 0, shanghai), want: 3},
		{name: "fdfsdf11", lastCheckIn: time.Date(2024, 11, 1, 23, 58, 0, 0, time.UTC), now: time.Date(2024, 11, 4, 23, 59, 0, 0, time.UTC), want: 3},
		{name: "ffsdzffs", lastCheckIn: time.Date(2024, 11, 2, 0, 2, 0, 0, shanghai), now: time.Date(2024, 11, 4, 23, 59, 0, 0, shanghai), want: 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateDaysDiff(tt.lastCheckIn, tt.now); got != tt.want {
				t.Errorf("calculateDaysDiff() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewEmptySignInRecorder(t *testing.T) {
	loc := time.UTC
	recorder := NewEmptySignInRecorder(loc)

	if recorder.record != 0 {
		t.Errorf("Expected record to be 0, got %d", recorder.record)
	}

	if recorder.lastSignInTime.IsZero() == false {
		t.Errorf("Expected lastSignInTime to be zero value time, got %v", recorder.lastSignInTime)
	}
	if recorder.lastSignInTime.Location() != loc {
		t.Errorf("Expected lastSignInTime location to be %v, got %v", loc, recorder.lastSignInTime.Location())
	}

	if recorder.location != loc {
		t.Errorf("Expected location to be %v, got %v", loc, recorder.location)
	}

	if recorder.clock == nil {
		t.Error("Expected clock to be initialized, but it's nil")
	}
}

func TestNewSignInRecorder(t *testing.T) {
	loc := time.UTC
	now := time.Now().In(loc)

	// Test cases for different Integer types
	testCases := []struct {
		name           string
		recordRaw      any
		lastSignInTime time.Time
		expectedRecord uint64
	}{
		{
			name:           "uint64",
			recordRaw:      uint64(0b1010),
			lastSignInTime: now,
			expectedRecord: uint64(0b1010),
		},
		{
			name:           "uint32",
			recordRaw:      uint32(0b1100),
			lastSignInTime: now,
			expectedRecord: uint64(0b1100),
		},
		{
			name:           "uint16",
			recordRaw:      uint16(0b1111),
			lastSignInTime: now,
			expectedRecord: uint64(0b1111),
		},
		{
			name:           "uint8",
			recordRaw:      uint8(0b0001),
			lastSignInTime: now,
			expectedRecord: uint64(0b0001),
		},
		{
			name:           "int64",
			recordRaw:      int64(0b1010),
			lastSignInTime: now,
			expectedRecord: uint64(0b1010),
		},
		{
			name:           "int64_negative",
			recordRaw:      int64(-0b1010),
			lastSignInTime: now,
			expectedRecord: uint64(0xfffffffffffffff6),
		},
		{
			name:           "int32",
			recordRaw:      int32(0b1100),
			lastSignInTime: now,
			expectedRecord: uint64(0b1100),
		},
		{
			name:           "int32_negative",
			recordRaw:      int32(-0b1100),
			lastSignInTime: now,
			expectedRecord: uint64(0b11111111111111111111111111110100),
		},
		{
			name:           "int16",
			recordRaw:      int16(0b1111),
			lastSignInTime: now,
			expectedRecord: uint64(0b1111),
		},
		{
			name:           "int16_negative",
			recordRaw:      int16(-0b1111),
			lastSignInTime: now,
			expectedRecord: uint64(0xfff1),
		},
		{
			name:           "int8",
			recordRaw:      int8(0b0001),
			lastSignInTime: now,
			expectedRecord: uint64(0b0001),
		},
		{
			name:           "int8_negative",
			recordRaw:      int8(-0b0001),
			lastSignInTime: now,
			expectedRecord: uint64(0xff),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			recorder := newSignInRecorderForTest(tc.recordRaw, tc.lastSignInTime)
			if recorder.record != tc.expectedRecord {
				t.Errorf("For %s: Expected record to be %b, got %b", tc.name, tc.expectedRecord, recorder.record)
			}
			if !recorder.lastSignInTime.Equal(tc.lastSignInTime) {
				t.Errorf("For %s: Expected lastSignInTime to be %v, got %v", tc.name, tc.lastSignInTime, recorder.lastSignInTime)
			}
			if recorder.location != loc {
				t.Errorf("For %s: Expected location to be %v, got %v", tc.name, loc, recorder.location)
			}
			if recorder.clock == nil {
				t.Errorf("For %s: Expected clock to be initialized, but it's nil", tc.name)
			}
		})
	}
}

func TestNewSignInRecorder2(t *testing.T) {
	loc := time.UTC
	// Test case 1: New empty recorder
	recorder := NewEmptySignInRecorder(loc)
	if recorder.record != 0 {
		t.Errorf("NewEmptySignInRecorder: expected record to be 0, got %d", recorder.record)
	}
	if !recorder.lastSignInTime.IsZero() {
		t.Errorf("NewEmptySignInRecorder: expected lastSignInTime to be zero, got %v", recorder.lastSignInTime)
	}
	if recorder.location != loc {
		t.Errorf("NewEmptySignInRecorder: expected location to be %v, got %v", loc, recorder.location)
	}

	// Test case 2: New recorder with uint64 record
	now := time.Now()
	recorder = NewSignInRecorder[uint64](1, now)
	if recorder.record != 1 {
		t.Errorf("NewSignInRecorder[uint64]: expected record to be 1, got %d", recorder.record)
	}
	if !recorder.lastSignInTime.Equal(now) {
		t.Errorf("NewSignInRecorder[uint64]: expected lastSignInTime to be %v, got %v", now, recorder.lastSignInTime)
	}
	if recorder.location != now.Location() {
		t.Errorf("NewSignInRecorder[uint64]: expected location to be %v, got %v", now.Location(), recorder.location)
	}

	// Test case 3: New recorder with int64 record
	recorder = NewSignInRecorder[int64](-1, now)
	if recorder.record != 18446744073709551615 {
		t.Errorf("NewSignInRecorder[int64]: expected record to be -1 as uint64, got %d", recorder.record)
	}
	if !recorder.lastSignInTime.Equal(now) {
		t.Errorf("NewSignInRecorder[int64]: expected lastSignInTime to be %v, got %v", now, recorder.lastSignInTime)
	}
	if recorder.location != now.Location() {
		t.Errorf("NewSignInRecorder[int64]: expected location to be %v, got %v", now.Location(), recorder.location)
	}

	recorder = NewSignInRecorder[int32](-1, now)
	if recorder.record != 4294967295 {
		t.Errorf("NewSignInRecorder[int32]: expected record to be -1 as uint64, got %d", recorder.record)
	}
	recorder = NewSignInRecorder[int16](-1, now)
	if recorder.record != 65535 {
		t.Errorf("NewSignInRecorder[int16]: expected record to be -1 as uint64, got %d", recorder.record)
	}
	recorder = NewSignInRecorder[int8](-1, now)
	if recorder.record != 255 {
		t.Errorf("NewSignInRecorder[int8]: expected record to be -1 as uint64, got %d", recorder.record)
	}

}

func TestSignIn(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)

	// Test case 1: First sign-in
	recorder := NewEmptySignInRecorder(loc)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()
	if recorder.record != 1 {
		t.Errorf("SignIn: expected record to be 1 after first sign-in, got %d", recorder.record)
	}
	if !recorder.lastSignInTime.Equal(mockTime) {
		t.Errorf("SignIn: expected lastSignInTime to be %v, got %v", mockTime, recorder.lastSignInTime)
	}
	// Test case 2: Second sign-in on the same day
	recorder.SignIn()
	if recorder.record != 1 {
		t.Errorf("SignIn: expected record to be still 1 after second sign-in, got %d", recorder.record)
	}

	// Test case 3: Sign-in next day
	mockTime = mockTime.AddDate(0, 0, 1)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()
	if recorder.record != 3 {
		t.Errorf("SignIn: expected record to be 3, got %d", recorder.record)
	}
	if !recorder.lastSignInTime.Equal(mockTime) {
		t.Errorf("SignIn: expected lastSignInTime to be %v, got %v", mockTime, recorder.lastSignInTime)
	}

	// Test case 4: sign-in after 65 days
	mockTime = mockTime.AddDate(0, 0, 65)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()
	if recorder.record != 1 {
		t.Errorf("SignIn: expected record to be reset, got %d", recorder.record)
	}
	if !recorder.lastSignInTime.Equal(mockTime) {
		t.Errorf("SignIn: expected lastSignInTime to be %v, got %v", mockTime, recorder.lastSignInTime)
	}
}

func TestCorrectSignInRecord(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)

	recorder := NewEmptySignInRecorder(loc)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()

	// Test case 1: same day
	sameDay := mockTime.Add(time.Hour)
	signed := recorder.correctSignInRecord(sameDay)
	if !signed {
		t.Errorf("correctSignInRecord: expected true for same day, got false")
	}
	if recorder.record != 1 {
		t.Errorf("correctSignInRecord: expected record is 1, got %d", recorder.record)
	}

	// Test case 2: next day
	nextDay := mockTime.AddDate(0, 0, 1)
	signed = recorder.correctSignInRecord(nextDay)
	if signed {
		t.Errorf("correctSignInRecord: expected false for next day, got true")
	}
	if recorder.record != 2 {
		t.Errorf("correctSignInRecord: expected record to be shifted, got %d", recorder.record)
	}

	// Test case 3: two days later
	twoDaysLater := mockTime.AddDate(0, 0, 2)
	signed = recorder.correctSignInRecord(twoDaysLater)
	if signed {
		t.Errorf("correctSignInRecord: expected false for 2 days later, got true")
	}
	if recorder.record != 8 {
		t.Errorf("correctSignInRecord: expected record to be 8, got %d", recorder.record)
	}

	// Test case 4: beyond 64 days
	recorder.record = 1
	recorder.lastSignInTime = mockTime
	beyond64Days := mockTime.AddDate(0, 0, 65)
	signed = recorder.correctSignInRecord(beyond64Days)
	if signed {
		t.Errorf("correctSignInRecord: expected false after 65 days, got true")
	}
	if recorder.record != 0 {
		t.Errorf("correctSignInRecord: expected record to be reset after 65 days, got %d", recorder.record)
	}
}

func TestCalculateDaysDiff(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)
	// Test case 1: same day
	diff := calculateDaysDiff(mockTime, mockTime.Add(5*time.Hour))
	if diff != 0 {
		t.Errorf("calculateDaysDiff: expected 0 for same day, got %d", diff)
	}

	// Test case 2: next day
	nextDay := mockTime.AddDate(0, 0, 1)
	diff = calculateDaysDiff(mockTime, nextDay)
	if diff != 1 {
		t.Errorf("calculateDaysDiff: expected 1 for next day, got %d", diff)
	}

	// Test case 3: next day with time difference
	diff = calculateDaysDiff(mockTime, nextDay.Add(10*time.Hour))
	if diff != 1 {
		t.Errorf("calculateDaysDiff: expected 1 for next day with time difference, got %d", diff)
	}

	// Test case 4: two days later
	twoDaysLater := mockTime.AddDate(0, 0, 2)
	diff = calculateDaysDiff(mockTime, twoDaysLater)
	if diff != 2 {
		t.Errorf("calculateDaysDiff: expected 2 for two days later, got %d", diff)
	}

	// Test case 5: Daylight saving time change (assuming DST starts on March)
	dstStart := time.Date(2024, time.March, 10, 2, 0, 0, 0, loc)
	beforeDST := dstStart.Add(-time.Hour * 3)
	diff = calculateDaysDiff(beforeDST, dstStart)
	if diff != 1 {
		t.Errorf("calculateDaysDiff: expected 1 for DST start day, got %d", diff)
	}

	dstEnd := time.Date(2024, time.November, 3, 1, 0, 0, 0, loc)
	beforeDSTEnd := dstEnd.Add(-time.Hour * 3)
	diff = calculateDaysDiff(beforeDSTEnd, dstEnd)
	if diff != 1 {
		t.Errorf("calculateDaysDiff: expected 1 for DST end day, got %d", diff)
	}

}

func TestRawRecord(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)
	recorder := NewEmptySignInRecorder(loc)
	recorder.SetClock(func() time.Time { return mockTime })

	// Test case 1: no sign-in
	record, lastSignInTime := recorder.RawRecord()
	if record != 0 {
		t.Errorf("RawRecord: expected record to be 0, got %d", record)
	}
	if !lastSignInTime.IsZero() {
		t.Errorf("RawRecord: expected lastSignInTime to be zero, got %v", lastSignInTime)
	}

	// Test case 2: after one sign-in
	recorder.SignIn()
	record, lastSignInTime = recorder.RawRecord()
	if record != 1 {
		t.Errorf("RawRecord: expected record to be 1, got %d", record)
	}
	if !lastSignInTime.Equal(mockTime) {
		t.Errorf("RawRecord: expected lastSignInTime to be %v, got %v", mockTime, lastSignInTime)
	}

	// Test case 3: after multiple sign-ins
	mockTime = mockTime.AddDate(0, 0, 2)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()
	record, lastSignInTime = recorder.RawRecord()
	if record != 5 {
		t.Errorf("RawRecord: expected record to be 5, got %d", record)
	}
	if !lastSignInTime.Equal(mockTime) {
		t.Errorf("RawRecord: expected lastSignInTime to be %v, got %v", mockTime, lastSignInTime)
	}
}

func TestRecord(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)
	recorder := NewEmptySignInRecorder(loc)
	recorder.SetClock(func() time.Time { return mockTime })

	// Test case 1: no sign-in
	record := recorder.Record()
	if len(record) != 64 {
		t.Errorf("Record: expected record length to be 64, got %d", len(record))
	}
	for _, signed := range record {
		if signed {
			t.Errorf("Record: expected all values to be false, got true")
		}
	}
	// Test case 2: one sign-in
	recorder.SignIn()
	record = recorder.Record()
	if !record[0] {
		t.Errorf("Record: expected record[0] to be true after first sign-in, got false")
	}
	for i := 1; i < len(record); i++ {
		if record[i] {
			t.Errorf("Record: expected record[%d] to be false, got true", i)
		}
	}

	// Test case 3: multiple sign-ins
	mockTime = mockTime.AddDate(0, 0, 2)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()
	record = recorder.Record()
	if !record[0] || !record[2] {
		t.Errorf("Record: expected record[0] and record[2] to be true, got false")
	}
	for i := 1; i < len(record); i++ {
		if i == 2 {
			continue
		}
		if record[i] {
			t.Errorf("Record: expected record[%d] to be false, got true", i)
		}
	}
}

func TestGetSignInRecordN(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)
	recorder := NewEmptySignInRecorder(loc)
	recorder.SetClock(func() time.Time { return mockTime })

	// Test case 1: valid n
	n := 10
	record := recorder.RecordN(n)
	if len(record) != n {
		t.Errorf("GetSignInRecordN: expected length to be %d, got %d", n, len(record))
	}

	// Test case 2: n = 0
	n = 0
	record = recorder.RecordN(n)
	if len(record) != 0 {
		t.Errorf("GetSignInRecordN: expected length to be %d, got %d", n, len(record))
	}
	// Test case 3: sign in and check
	recorder.SignIn()
	record = recorder.RecordN(1)
	if !record[0] {
		t.Errorf("GetSignInRecordN: expected index 0 to be true after signing in, got false")
	}

	// Test case 4: invalid n
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("GetSignInRecordN: expected panic for n out of range")
		}
	}()
	recorder.RecordN(65)
}

func TestHasSignIn(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)
	recorder := NewEmptySignInRecorder(loc)
	recorder.SetClock(func() time.Time { return mockTime })

	// Test case 1: no sign-in
	if recorder.HasSignIn(0) {
		t.Errorf("HasSignIn: expected false for no sign-in, got true")
	}

	// Test case 2: valid day sign-in
	recorder.SignIn()
	if !recorder.HasSignIn(0) {
		t.Errorf("HasSignIn: expected true for sign-in on day 0, got false")
	}
	if recorder.HasSignIn(1) {
		t.Errorf("HasSignIn: expected false for not sign-in on day 1, got true")
	}
	// Test case 3: invalid day
	if recorder.HasSignIn(-1) {
		t.Errorf("HasSignIn: expected false for negative day index, got true")
	}
	if recorder.HasSignIn(64) {
		t.Errorf("HasSignIn: expected false for day index over 63, got true")
	}
	// Test case 4: multi-sign-in
	mockTime = mockTime.AddDate(0, 0, 2)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()
	if !recorder.HasSignIn(2) {
		t.Errorf("HasSignIn: expected true for sign-in on day 2, got false")
	}

}

func TestHasSignedToday(t *testing.T) {
	loc, _ := time.LoadLocation("America/New_York")
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)
	recorder := NewEmptySignInRecorder(loc)
	recorder.SetClock(func() time.Time { return mockTime })

	// Test case 1: first sign-in of the day
	if recorder.HasSignedToday() {
		t.Errorf("HasSignedToday: expected false before sign-in, got true")
	}
	recorder.SignIn()
	if !recorder.HasSignedToday() {
		t.Errorf("HasSignedToday: expected true after first sign-in, got false")
	}

	// Test case 2: sign-in next day
	mockTime = mockTime.AddDate(0, 0, 1)
	recorder.SetClock(func() time.Time { return mockTime })
	if recorder.HasSignedToday() {
		t.Errorf("HasSignedToday: expected false before sign-in, got true")
	}
	recorder.SignIn()
	if !recorder.HasSignedToday() {
		t.Errorf("HasSignedToday: expected true after first sign-in, got false")
	}

	// Test case 3: same day sign-in
	if !recorder.HasSignedToday() {
		t.Errorf("HasSignedToday: expected true after same day sign-in, got false")
	}
}

func TestSignInWithDifferentTimeZones(t *testing.T) {
	utc := time.UTC
	loc, _ := time.LoadLocation("America/Los_Angeles") // PST
	mockTime := time.Date(2024, time.January, 1, 10, 0, 0, 0, loc)

	// Create two recorders, one in UTC, one in PST
	recorderUTC := NewEmptySignInRecorder(utc)
	recorderUTC.SetClock(func() time.Time { return mockTime.In(utc) })
	recorderPST := NewEmptySignInRecorder(loc)
	recorderPST.SetClock(func() time.Time { return mockTime })

	// First sign-in.
	recorderUTC.SignIn()
	recorderPST.SignIn()

	// Check raw record
	rawRecordUTC, _ := recorderUTC.RawRecord()
	rawRecordPST, _ := recorderPST.RawRecord()
	if rawRecordUTC != 1 {
		t.Errorf("TestSignInWithDifferentTimeZones (UTC): Expected record to be 1, got %d", rawRecordUTC)
	}
	if rawRecordPST != 1 {
		t.Errorf("TestSignInWithDifferentTimeZones (PST): Expected record to be 1, got %d", rawRecordPST)
	}

	// Move to next day in PST
	mockTime = mockTime.AddDate(0, 0, 1)
	recorderUTC.SetClock(func() time.Time { return mockTime.In(utc) })
	recorderPST.SetClock(func() time.Time { return mockTime })

	//Second sign-in
	recorderUTC.SignIn()
	recorderPST.SignIn()

	// check raw record again
	rawRecordUTC, _ = recorderUTC.RawRecord()
	rawRecordPST, _ = recorderPST.RawRecord()
	if rawRecordUTC != 3 {
		t.Errorf("TestSignInWithDifferentTimeZones (UTC): Expected record to be 3, got %d", rawRecordUTC)
	}
	if rawRecordPST != 3 {
		t.Errorf("TestSignInWithDifferentTimeZones (PST): Expected record to be 3, got %d", rawRecordPST)
	}

	// Check HasSignIn
	if !recorderUTC.HasSignIn(0) {
		t.Errorf("TestSignInWithDifferentTimeZones (UTC): Expected HasSignIn(0) to be true, got false")
	}
	if !recorderPST.HasSignIn(0) {
		t.Errorf("TestSignInWithDifferentTimeZones (PST): Expected HasSignIn(0) to be true, got false")
	}
	if !recorderUTC.HasSignIn(1) {
		t.Errorf("TestSignInWithDifferentTimeZones (UTC): Expected HasSignIn(1) to be true, got false")
	}
	if !recorderPST.HasSignIn(1) {
		t.Errorf("TestSignInWithDifferentTimeZones (PST): Expected HasSignIn(1) to be true, got false")
	}

	// Check hasSignedToday
	if !recorderUTC.HasSignedToday() {
		t.Errorf("TestSignInWithDifferentTimeZones (UTC): Expected HasSignedToday() to be true, got false")
	}
	if !recorderPST.HasSignedToday() {
		t.Errorf("TestSignInWithDifferentTimeZones (PST): Expected HasSignedToday() to be true, got false")
	}
}

func TestSignInRecorder_GetConsecutiveSignInDays(t *testing.T) {
	loc := time.UTC
	now := time.Date(2024, 1, 10, 12, 0, 0, 0, loc)
	clock := func() time.Time { return now }
	tests := []struct {
		name           string
		record         uint64
		lastSignInTime time.Time
		expectedDays   int
	}{
		{
			name:           "empty record",
			record:         0,
			lastSignInTime: now.Add(-time.Hour * 24),
			expectedDays:   0,
		},
		{
			name:           "single sign-in",
			record:         1,
			lastSignInTime: now,
			expectedDays:   1,
		},
		{
			name:           "multiple consecutive sign-ins",
			record:         0b1111,
			lastSignInTime: now,
			expectedDays:   4,
		},
		{
			name:           "non-consecutive sign-ins",
			record:         0b1011,
			lastSignInTime: now,
			expectedDays:   2,
		},
		{
			name:           "all 64 sign-ins",
			record:         ^uint64(0),
			lastSignInTime: now,
			expectedDays:   64,
		},
		{
			name:           "sign-ins with time gap",
			record:         0b11,
			lastSignInTime: now.Add(-time.Hour * 25),
			expectedDays:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSignInRecorder(tt.record, tt.lastSignInTime.In(loc))
			s.SetClock(clock)
			actualDays := s.ConsecutiveSignInDays()
			if actualDays != tt.expectedDays {
				t.Errorf("GetConsecutiveSignInDays() = %v, want %v", actualDays, tt.expectedDays)
			}
		})
	}
}

func TestSignInRecorder_DaysToNextMilestone(t *testing.T) {
	loc := time.UTC
	now := time.Date(2024, 1, 10, 12, 0, 0, 0, loc)
	clock := func() time.Time { return now }
	tests := []struct {
		name           string
		record         uint64
		lastSignInTime time.Time
		milestone      int
		expectedDays   int
	}{
		{
			name:           "empty record, milestone 7",
			record:         0,
			lastSignInTime: now,
			milestone:      7,
			expectedDays:   7,
		},
		{
			name:           "empty record, milestone 30",
			record:         0,
			lastSignInTime: now,
			milestone:      30,
			expectedDays:   30,
		},
		{
			name:           "single sign-in, milestone 7",
			record:         1,
			lastSignInTime: now,
			milestone:      7,
			expectedDays:   6,
		},
		{
			name:           "consecutive sign-ins < milestone",
			record:         0b111,
			lastSignInTime: now,
			milestone:      7,
			expectedDays:   4,
		},
		{
			name:           "consecutive sign-ins == milestone",
			record:         0b1111111,
			lastSignInTime: now,
			milestone:      7,
			expectedDays:   0,
		},
		{
			name:           "consecutive sign-ins > milestone",
			record:         0b11111111,
			lastSignInTime: now,
			milestone:      7,
			expectedDays:   0,
		},
		{
			name:           "consecutive sign-ins with time gap",
			record:         0b11,
			lastSignInTime: now.Add(-time.Hour * 25),
			milestone:      7,
			expectedDays:   7,
		},
		{
			name:           "consecutive sign-ins with large milestone",
			record:         0b1111111,
			lastSignInTime: now,
			milestone:      100,
			expectedDays:   93,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSignInRecorder(tt.record, tt.lastSignInTime.In(loc))
			s.SetClock(clock)
			actualDays := s.DaysToNextMilestone(tt.milestone)
			if actualDays != tt.expectedDays {
				t.Errorf("DaysToNextMilestone(%d) = %v, want %v", tt.milestone, actualDays, tt.expectedDays)
			}
		})
	}
}
func ExampleSignInRecorder() {
	loc, _ := time.LoadLocation("America/New_York")
	recorder := NewEmptySignInRecorder(loc)

	fmt.Println("Initial record:", recorder.Record())

	recorder.SignIn()

	fmt.Println("Record after first sign-in:", recorder.Record())

	recorder.SignIn()
	fmt.Println("Has signed today:", recorder.HasSignedToday())
	fmt.Println("Record after same day sign-in:", recorder.Record())

	mockTime := time.Now().AddDate(0, 0, 2)
	recorder.SetClock(func() time.Time { return mockTime })
	recorder.SignIn()

	fmt.Println("Record after 2 days later sign-in:", recorder.Record())
	fmt.Println("Sign-in on day 0:", recorder.HasSignIn(0))
	fmt.Println("Sign-in on day 2:", recorder.HasSignIn(2))

	rawRecord, _ := recorder.RawRecord()
	fmt.Println("Raw record:", rawRecord)

	n := 3
	nRecord := recorder.RecordN(n)
	fmt.Println("Get record with n = ", n, ":", nRecord)
	// Output:
	// Initial record: [false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false]
	// Record after first sign-in: [true false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false]
	// Has signed today: true
	// Record after same day sign-in: [true false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false]
	// Record after 2 days later sign-in: [true false true false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false false]
	// Sign-in on day 0: true
	// Sign-in on day 2: true
	// Raw record: 5
	// Get record with n =  3 : [true false true]
}
