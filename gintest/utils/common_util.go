package my_utils

import (
	"unsafe"
)

// 将string转为[]byte
func StringToBytes(s string) []byte {
	// 注意该方式返回的byte不可更改
	strHeader := unsafe.StringData(s)
	byteSlice := unsafe.Slice(strHeader, len(s))
	return byteSlice
}
