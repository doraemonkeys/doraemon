package doraemon

import (
	"errors"
	"math/big"
	"strconv"
	"strings"
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
