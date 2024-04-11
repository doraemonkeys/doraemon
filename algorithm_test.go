package doraemon

import (
	"errors"
	"testing"
)

func TestHexToInt2(t *testing.T) {
	tests := []struct {
		hex      string
		expected int
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
