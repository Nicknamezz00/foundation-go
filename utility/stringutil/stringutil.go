package stringutil

import (
	"strconv"
	"unsafe"
)

func ToBytes(str string) []byte {
	x := (*[2]uintptr)(unsafe.Pointer(&str))
	h := [3]uintptr{x[0], x[1], x[1]}
	return *(*[]byte)(unsafe.Pointer(&h))
}

func FromBytes(bytes []byte) string {
	return *(*string)(unsafe.Pointer(&bytes))
}

func ToInt64(str string) (int64, error) {
	return strconv.ParseInt(str, 10, 64)
}
