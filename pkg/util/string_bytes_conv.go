package util

import "unsafe"

// String2Bytes 将字符串转换为字节切片
// 注意：返回的字节切片和 str 共享内存，因此最好不要修改返回的字节切片
func String2Bytes(str string) []byte {
	if str == "" {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(str), len(str))
}

// Bytes2String 将字节切片转换为字符串
// 注意：返回的字符串和 bs 共享内存，因此一定不要修改返回的字符串！
func Bytes2String(bs []byte) string {
	if len(bs) == 0 {
		return ""
	}
	return unsafe.String(unsafe.SliceData(bs), len(bs))
}
