package doraemon

import "unsafe"

func StringToReadOnlyBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// BytesToString converts a byte slice to a string without copying.
// The byte slice must not be modified after the conversion.
// otherwise, the string may be corrupted.
func BytesToString(b []byte) string {
	return unsafe.String(&b[0], len(b))
}
