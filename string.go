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

// SplitBySpaceLimit2 splits a string by a minimum number of consecutive spaces.
//
// It takes a string `line` and an integer `spaceLimit` as input.
// It returns a slice of strings, where each string is a part of the original
// string that was separated by at least `spaceLimit + 1` spaces.  Leading and
// trailing spaces on each split part are trimmed.  Empty strings resulting from
// the split are not included in the output, unless all resulting splits are empty.
//
// Example:
//
//	SplitBySpaceLimit2("  foo   bar    baz", 3) == ["foo   bar", "baz"]
//	SplitBySpaceLimit2("foo       bar", 3)  == ["foo", "bar"]
//	SplitBySpaceLimit2("   foo   ", 2) == ["foo"]
//	SplitBySpaceLimit2("      ", 2) == []
//	SplitBySpaceLimit2("foo", 2) == ["foo"]
func SplitBySpaceLimit2(line string, spaceLimit int) []string {
	var minSpaceStr = strings.Repeat(" ", spaceLimit+1)
	var values []string
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

// ScanFields parses a set of text lines (lines) and extracts the field names (fields) from the first line.
// It then parses the remaining lines as records, where each record is a collection of field values.
// This function handles a limited number of space characters to separate fields in the records.
//
// Parameters:
//   - lines: A slice of strings containing the text lines to be parsed.
//   - fieldSpaceLimit: An integer specifying the minimum number of spaces between fields to consider them separate.
//     If 0, any number of spaces is used to separate fields.
//   - recordSpaceLimit: An integer specifying the minimum number of spaces between values in a record to consider them separate.
//     If 0, any number of spaces is used to separate values.
//   - fieldKeyWords:  A map used for fine-grained field association.  Keys are field names,
//     and values are functions that take a string (a potential field value) and return true if
//     the value matches that field, helping disambiguate in cases of misalignment.
//
// It is guaranteed that the first line represents field names, and each subsequent line is a valid record.
//
// Returns:
//   - fields: A slice of strings representing the names of the fields.
//   - records: A slice of string slices, where each inner slice represents a record and contains the values for each field.
//   - err: An error object, which is non-nil if an error occurs during parsing (e.g., empty input, no fields found, or an unsupported line format).
//
// Example:
//
//	Input lines:
//	  Name Access  Availability  BlockSize
//	  C:     3       0           4096
//	  D:     3                  4096
//	  E:     3       1           4096
//	  F:            1           4096
//
//	Output:
//	  fields = ["Name", "Access", "Availability", "BlockSize"]
//	  records = [
//	    ["C:", "3", "0", "4096"],
//	    ["D:", "3", "", "4096"],
//	    ["E:", "3", "1", "4096"],
//	    ["F:", "", "1", "4096"],
//	  ]
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

	// Create a map to store the index (position) of each field.
	var fieldsIndexMap = make(map[string]int, len(fields))
	scanCount := 0
	// Calculate and store the index of each field, considering potential Unicode characters (e.g., Chinese).
	for _, field := range fields {
		asciiIndex := scanCount + strings.Index(lines[0][scanCount:], field)
		scanCount = asciiIndex + len(field)
		runeLen := len([]rune(lines[0][:asciiIndex]))
		// 两个汉字大约多占用一个空格
		// Estimate additional space occupied by wide characters (like Chinese).
		fieldsIndexMap[field] = runeLen + (asciiIndex-runeLen)/4
	}

	// Iterate through the remaining lines to parse records.
	for _, line := range lines[1:] {
		var record = make([]string, len(fields))
		var values []string

		// Split the record line into values based on recordSpaceLimit.
		if recordSpaceLimit == 0 {
			values = strings.Fields(line)
		} else {
			values = SplitBySpaceLimit2(line, recordSpaceLimit)
		}

		// Easy case: If the number of values matches the number of fields, directly copy the values.
		if len(values) == len(fields) {
			copy(record, values)
			records = append(records, record)
			continue
		}

		// Complex case: Handle lines where the number of values doesn't match the number of fields (due to missing values).
		holesWithNothingI := 0 // Keep track of the field index where a value might be missing.
		scanCount = 0
		for valueI, value := range values {
			// Find the character position of the current value in the line.
			valueAsciiIndex := scanCount + strings.Index(line[scanCount:], value)
			scanCount = valueAsciiIndex + len(value)

			// Calculate the rune index and adjust for potential wide characters.
			valueRuneIndex := len([]rune(line[:valueAsciiIndex]))
			valueCalibration := (valueAsciiIndex - valueRuneIndex) / 4
			valueIndex := valueRuneIndex + valueCalibration // 修正后的长度索引 Corrected index.

			// Calculate the index for the tail of the current value
			valueTailAsciiIndex := valueAsciiIndex + len(value)
			valueTailRuneIndex := len([]rune(line[:valueTailAsciiIndex]))
			valueTailCalibration := (valueTailAsciiIndex - valueTailRuneIndex) / 4
			valueTailIndex := valueTailRuneIndex + valueTailCalibration

			// Calculate a center index of the value. It used for march
			valueCenterIndex := valueIndex + (valueTailIndex-valueIndex)/2

			// Determine the correct field index (holeIndex) for the current value.
			holeIndex := matchValueIndex(value, fieldKeyWords, fieldsIndexMap,
				fields, holesWithNothingI, valueCenterIndex)
			holesWithNothingI = holeIndex + 1 // Update the starting index for the next value.

			// If all the remain value can not match the field, it is an unsupported line
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

// matchValueIndex determines the correct field index for a given value, considering potential misalignments
// and using field keywords for disambiguation.
func matchValueIndex(
	value string,
	fieldKeyWords map[string]func(string) bool,
	fieldsIndexMap map[string]int,
	fields []string,
	holesWithNothingI int,
	valueCenterIndex int) (holeIndex int) {

	// Handle the first field (or single-field case).
	if holesWithNothingI == 0 {
		if len(fields) == 1 {
			return holesWithNothingI
		}
		if valueCenterIndex < fieldsIndexMap[fields[1]] {
			return holesWithNothingI // Value belongs to the first field.
		}
		holesWithNothingI = 1 // Otherwise, start checking from the second field.
	}

	// Iterate through the fields to find the best match.
	for i := holesWithNothingI; i < len(fields)-1; i++ {
		// Check if the value's center index falls within the range of the current field.
		if valueCenterIndex < fieldsIndexMap[fields[i+1]] {
			// Use field keywords for more accurate matching, if available.
			filter, exist := fieldKeyWords[fields[i]]
			filter_l, exist_l := fieldKeyWords[fields[i-1]]
			filter_r, exist_r := fieldKeyWords[fields[i+1]]

			// Check if the value matches the current, previous, or next field based on keywords.
			if exist && filter(value) {
				return i
			}
			if exist_l && filter_l(value) {
				return i - 1
			}
			if exist_r && filter_r(value) {
				return i + 1
			}
			return i // Return the current field index if no keywords match.
		}
	}
	// 默认为最后一个值对应最后一个字段
	return len(fields) - 1
}

// ReadLines reads all lines from the given io.Reader and returns them as a slice of byte slices.
// Each line has leading/trailing whitespace removed.
func ReadLines(reader io.Reader) (lines [][]byte, err error) {
	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, reader)
	if err != nil {
		return nil, err
	}
	lines = bytes.Split(buf.Bytes(), []byte("\n"))
	for i := 0; i < len(lines); i++ {
		lines[i] = bytes.TrimSpace(lines[i])
	}
	return lines, nil
}

// ReadTrimmedLines reads content from the given io.Reader, splits it into lines,
// and returns a slice of byte slices.  Leading and trailing empty lines (after whitespace trimming) are removed.
func ReadTrimmedLines(reader io.Reader) (lines [][]byte, err error) {
	lines, err = ReadLines(reader)
	if err != nil {
		return nil, err
	}

	// Remove leading empty lines.
	for i := 0; i < len(lines); i++ {
		if len(bytes.TrimSpace(lines[i])) != 0 {
			lines = lines[i:]
			break
		}
	}

	// Remove trailing empty lines.
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

// GenRandomAsciiString generate a random string of a specified length, only containing uppercase and lowercase letters and numbers.
func GenRandomAsciiString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return GenRandomString(charset, length)
}

func HasAnyPrefix(s string, prefixs ...string) bool {
	for _, prefix := range prefixs {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
