package doraemon

import (
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"strings"
	"unsafe"
)

func StringToBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts a byte slice to a string without copying.
// The byte slice must not be modified after the conversion.
// otherwise, the string may be corrupted.
func BytesToString(b []byte) string {
	return unsafe.String(&b[0], len(b))
}

// StrLen统计字符串长度
func StrLen(str string) int {
	return len([]rune(str))
}

func StringAsPointer(s string) (unsafe.Pointer, int) {
	data := unsafe.StringData(s)
	return unsafe.Pointer(data), len(s)
}

func BytesAsPointer(b []byte) (unsafe.Pointer, int, int) {
	return unsafe.Pointer(&b[0]), len(b), cap(b)
}

func GetCStringLen(str unsafe.Pointer) int {
	var i int
	for ; *(*byte)(unsafe.Pointer(uintptr(str) + uintptr(i))) != 0; i++ {
	}
	return i
}

func SplitBySpaceLimit(line string, spaceLimit int) []string {
	var values []string
	var valueBuf bytes.Buffer
	spaceCount := 0
	line = strings.TrimSpace(line)
	for _, r := range line {
		if r == ' ' {
			spaceCount++
			if spaceCount > spaceLimit {
				if valueBuf.Len() == 0 {
					continue
				}
				str := strings.TrimSpace(valueBuf.String())
				if str != "" {
					values = append(values, str)
				}
				valueBuf.Reset()
				spaceCount = 0
				continue
			}
		} else {
			spaceCount = 0
		}
		valueBuf.WriteRune(r)
	}
	if valueBuf.Len() > 0 {
		last := strings.TrimSpace(valueBuf.String())
		if last != "" {
			values = append(values, last)
		}
	}
	return values
}

func SplitBySpaceLimit2(line string, spaceLimit int) []string {
	var minSpaceStr = strings.Repeat(" ", spaceLimit+1)
	var values = make([]string, 0, 10)
	for {
		before, after, found := strings.Cut(line, minSpaceStr)
		if !found {
			lineVal := strings.TrimSpace(line)
			if lineVal != "" {
				values = append(values, lineVal)
			}
			return values
		}
		line = after
		beforeVal := strings.TrimSpace(before)
		if beforeVal == "" {
			continue
		}
		values = append(values, strings.TrimSpace(before))
	}
}

// ScanFields 函数的作用是解析一组文本行（lines），将第一行作为字段名称（fields）提取，
// 并将剩余的行解析为记录（records），每条记录是字段值的集合。
// 这个函数可以处理有限数量的空格字符来分隔记录中的字段。
// 函数的参数：
// lines: 一个字符串切片，包含要解析的文本行。
// recordSpaceLimit: 一个整数，指定字段之间的最小空格数来认为字段是分开的。
// 如果为0，则使用任意数量的空格来分隔字段。
// 需要保证第一行是字段名, 且后面每行都是有效的记录。
//
// # 例如，下面的文本行：
//
//	Name Access  Availability  BlockSize
//
// C:     3       0           4096
//
// D:     3                  4096
//
// E:     3       1           4096
//
// F:            1           4096
//
// 将被解析为：
//
// fields = ["Name", "Access", "Availability", "BlockSize"]
//
//	 records = [
//		["C:", "3", "0", "4096"],
//		["D:", "3", "", "4096"],
//		["E:", "3", "1", "4096"],
//		["F:", "", "1", "4096"],
//
// ]
func ScanFields(
	lines []string,
	fieldSpaceLimit int,
	recordSpaceLimit int,
	fieldKeyWords map[string]func(string) bool,
) (fields []string, records [][]string, err error) {
	if recordSpaceLimit < 0 {
		recordSpaceLimit = 0
	}
	if fieldSpaceLimit < 0 {
		fieldSpaceLimit = 0
	}
	if len(lines) == 0 {
		return nil, nil, fmt.Errorf("empty lines")
	}
	if fieldKeyWords == nil {
		fieldKeyWords = make(map[string]func(string) bool)
	}
	if fieldSpaceLimit == 0 {
		fields = strings.Fields(lines[0])
	} else {
		fields = SplitBySpaceLimit2(lines[0], fieldSpaceLimit)
	}
	if len(fields) == 0 {
		return nil, nil, fmt.Errorf("empty fields")
	}
	var fieldsIndexMap = make(map[string]int, len(fields))
	scanCount := 0
	for _, field := range fields {
		asciiIndex := scanCount + strings.Index(lines[0][scanCount:], field)
		scanCount = asciiIndex + len(field)
		runeLen := len([]rune(lines[0][:asciiIndex]))
		// 两个汉字大约多占用一个空格
		fieldsIndexMap[field] = runeLen + (asciiIndex-runeLen)/4
	}
	for _, line := range lines[1:] {
		var record = make([]string, len(fields))
		var values []string
		if recordSpaceLimit == 0 {
			values = strings.Fields(line)
		} else {
			values = SplitBySpaceLimit2(line, recordSpaceLimit)
		}
		// easy case
		if len(values) == len(fields) {
			copy(record, values)
			records = append(records, record)
			continue
		}

		holesWithNothingI := 0
		scanCount = 0
		for valueI, value := range values {
			valueAsciiIndex := scanCount + strings.Index(line[scanCount:], value)
			scanCount = valueAsciiIndex + len(value)
			valueRuneIndex := len([]rune(line[:valueAsciiIndex]))
			valueCalibration := (valueAsciiIndex - valueRuneIndex) / 4
			valueIndex := valueRuneIndex + valueCalibration //修正后的长度索引
			// valueCenterIndex := valueAsciiIndex + len(value)/2
			valueTailAsciiIndex := valueAsciiIndex + len(value)
			valueTailRuneIndex := len([]rune(line[:valueTailAsciiIndex]))
			valueTailCalibration := (valueTailAsciiIndex - valueTailRuneIndex) / 4
			valueTailIndex := valueTailRuneIndex + valueTailCalibration
			valueCenterIndex := valueIndex + (valueTailIndex-valueIndex)/2

			holeIndex := marchValueIndex(value, fieldKeyWords, fieldsIndexMap,
				fields, holesWithNothingI, valueCenterIndex)
			holesWithNothingI = holeIndex + 1
			if holesWithNothingI == len(fields) && valueI != len(values)-1 {
				return nil, nil, fmt.Errorf("unsupported line: %s", line)
			}
			record[holeIndex] = value
			// fmt.Printf("record[%d](%s) = %s,nextI=%d\n", nextI-1, fields[nextI-1], value, nextI)
		}
		records = append(records, record)
	}
	return fields, records, nil
}

func marchValueIndex(
	value string,
	fieldKeyWords map[string]func(string) bool,
	fieldsIndexMap map[string]int,
	fields []string,
	holesWithNothingI int,
	valueCenterIndex int) (holeIndex int) {
	if holesWithNothingI == 0 {
		if len(fields) == 1 {
			return holesWithNothingI
		}
		if valueCenterIndex < fieldsIndexMap[fields[1]] {
			return holesWithNothingI
		}
		holesWithNothingI = 1
	}
	for i := holesWithNothingI; i < len(fields)-1; i++ {
		if valueCenterIndex < fieldsIndexMap[fields[i+1]] {
			// return i
			filter, exist := fieldKeyWords[fields[i]]
			filter_l, exist_l := fieldKeyWords[fields[i-1]]
			filter_r, exist_r := fieldKeyWords[fields[i+1]]
			if exist && filter(value) {
				// fmt.Println("filter:", fields[i], value)
				return i
			}
			if exist_l && filter_l(value) {
				// fmt.Println("filter_l:", fields[i-1], value)
				return i - 1
			}
			if exist_r && filter_r(value) {
				// fmt.Println("filter_r:", fields[i+1], value)
				return i + 1
			}
			return i
		}
	}
	// 默认为最后一个值对应最后一个字段
	return len(fields) - 1
}

func ReadLines(reader io.Reader) (lines [][]byte, err error) {
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	return bytes.Split(buf.Bytes(), []byte("\n")), nil
}

// ReadTrimmedLines 从给定的 io.Reader 中读取内容，并按行分割成字节切片。
// 开头和结尾的空白行将被去除。
// 返回值 lines 是一个二维字节切片，每个元素代表一行的内容。
func ReadTrimmedLines(reader io.Reader) (lines [][]byte, err error) {
	lines, err = ReadLines(reader)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(lines); i++ {
		if len(bytes.TrimSpace(lines[i])) != 0 {
			lines = lines[i:]
			break
		}
	}
	for i := len(lines) - 1; i >= 0; i-- {
		if len(bytes.TrimSpace(lines[i])) != 0 {
			lines = lines[:i+1]
			break
		}
	}
	return lines, nil
}

func GenRandomString(charset string, length int) string {
	strB := strings.Builder{}
	strB.Grow(length)
	var err error
	for i := 0; i < length; i++ {
		if err = strB.WriteByte(charset[rand.IntN(len(charset))]); err != nil {
			panic(err)
		}
	}
	return strB.String()
}

// GenRandomAsciiString 生成指定长度的随机字符串，只包含大小写字母和数字。
func GenRandomAsciiString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return GenRandomString(charset, length)
}
