package doraemon

import (
	"errors"
	"fmt"
	"strconv"
)

//十六进制转换为十进制
func HexToInt(hex string) (int, error) {
	if len(hex) > 2 {
		if string(hex[0:2]) == "0x" || string(hex[0:2]) == "0X" {
			hex = hex[2:]
		}
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

//字符转整型
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

//对传入的用户ID进行基本的验证(字符|长度等)
func UserIDTest(userID string) bool {
	if userID == "" {
		return false
	}
	for _, v := range userID {
		if v >= '0' && v <= '9' {
			continue
		}
		return false
	}
	return true
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

//对传入的密码进行基本的验证(字符|长度等)
func PasswordTest(password string) bool {
	if password == "" {
		return false
	}
	for i := 0; i < len(password); i++ {
		if password[i] == ' ' || password[i] == '\n' || password[i] == '\t' {
			return false
		}
	}
	return true
}

//对传入的用户名进行基本的验证(字符|长度等)
func UserNameTest(userName string) bool {
	if userName == "" {
		return false
	}
	for i := 0; i < len(userName); i++ {
		if userName[i] == ' ' || userName[i] == '\n' || userName[i] == '\t' {
			return false
		}
	}
	return true
}
