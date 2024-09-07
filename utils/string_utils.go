package utils

import (
	"unsafe"
)

// BytesToString 没有内存开销的转换
func BytesToString(b []byte) string {
	return unsafe.String(unsafe.SliceData(b), len(b))
}

// StringToBytes 没有内存开销的转换
func StringToBytes(s string) (b []byte) {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
