package utils

import "unsafe"

func Strings(data *[]byte) string {
	return *(*string)(unsafe.Pointer(data))
}
