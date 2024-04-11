package doraemon

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// 十六进制转换为十进制
func HexToInt2(hex string) (int, error) {
	if len(hex) > 2 && (strings.HasPrefix(hex, "0x") || strings.HasPrefix(hex, "0X")) {
		hex = hex[2:]
	}
	var result int
	for _, v := range hex {
		result *= 16
		switch {
		case v >= '0' && v <= '9':
			result += int(v - '0')
		case v >= 'a' && v <= 'f':
			result += int(v - 'a' + 10)
		case v >= 'A' && v <= 'F':
			result += int(v - 'A' + 10)
		default:
			return 0, errors.New("invalid hex string")
		}
	}
	return result, nil
}

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

// 字符转整型
func CharToInt(c byte) int {
	if c >= '0' && c <= '9' {
		return int(c - '0')
	}
	if c >= 'a' && c <= 'f' {
		return int(c-'a') + 10
	}
	if c >= 'A' && c <= 'F' {
		return int(c-'A') + 10
	}
	return 0
}

// Str2uint64 将字符串转换为uint64 使用前请确保传入的字符串是合法的
func Str2uint64(str string) uint64 {
	res, _ := strconv.ParseUint(str, 10, 64)
	return res
}

// Str2int64 将字符串转换为int64 使用前请确保传入的字符串是合法的
func Str2int64(str string) int64 {
	res, _ := strconv.ParseInt(str, 10, 64)
	return res
}

// Str2int32 将字符串转换为int32 使用前请确保传入的字符串是合法的
func Str2int32(str string) int32 {
	// res, _ := strconv.ParseInt(str, 10, 32)
	// return int32(res)
	var res int32 = 0
	fmt.Sscanf("str", "%d", &res)
	return res
}
