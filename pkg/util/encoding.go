package util

import (
	"bytes"
	"encoding/binary"
)

// WriteUint8 编码一个无符号 8 位整数到 buf 中
func WriteUint8(buf *bytes.Buffer, n uint8) error {
	return buf.WriteByte(n)
}

// WriteVarUint 编码一个无符号整数到 buf 中，
// 使用变长编码，每个字节的高位表示是否继续，低位表示数值
func WriteVarUint(buf *bytes.Buffer, x uint64) error {
	for x >= 0x80 {
		err := buf.WriteByte(byte(x) | 0x80)
		if err != nil {
			return err
		}
		x >>= 7
	}
	return buf.WriteByte(byte(x))
}

// WriteVarString 编码一个字符串到 buf 中
// 先写入字符串的长度，然后写入字符串
func WriteVarString(buf *bytes.Buffer, str string) error {
	if err := WriteVarUint(buf, uint64(len(str))); err != nil {
		return err
	}
	_, err := buf.WriteString(str)
	return err
}

// WriteUint32 编码一个无符号 32 位整数到 buf 中，使用小端编码，大小为 4 字节
func WriteUint32(buf *bytes.Buffer, n uint32) error {
	return binary.Write(buf, binary.LittleEndian, n)
}

// WriteUint64 编码一个无符号 64 位整数到 buf 中，使用小端编码，大小为 8 字节
func WriteUint64(buf *bytes.Buffer, n uint64) error {
	return binary.Write(buf, binary.LittleEndian, n)
}

// WriteInt8 编码一个有符号 8 位整数到 buf 中
func WriteInt8(buf *bytes.Buffer, n int8) error {
	return buf.WriteByte(byte(n))
}

// WriteInt16 编码一个有符号 16 位整数到 buf 中，使用小端编码
func WriteInt16(buf *bytes.Buffer, n int16) error {
	return binary.Write(buf, binary.LittleEndian, n)
}

// WriteInt32 编码一个有符号 32 位整数到 buf 中，使用小端编码
func WriteInt32(buf *bytes.Buffer, n int32) error {
	return binary.Write(buf, binary.LittleEndian, n)
}

// WriteInt64 编码一个有符号 64 位整数到 buf 中，使用小端编码
func WriteInt64(buf *bytes.Buffer, n int64) error {
	return binary.Write(buf, binary.LittleEndian, n)
}

// WriteFloat32 编码一个 32 位浮点数到 buf 中，使用小端编码
func WriteFloat32(buf *bytes.Buffer, n float32) error {
	return binary.Write(buf, binary.LittleEndian, n)
}

// WriteFloat64 编码一个 64 位浮点数到 buf 中，使用小端编码
func WriteFloat64(buf *bytes.Buffer, n float64) error {
	return binary.Write(buf, binary.LittleEndian, n)
}

// WriteVarByteArray 编码一个变长字节数组到 buf 中
// 先写入数组的长度，然后写入数组内容
func WriteVarByteArray(buf *bytes.Buffer, b []byte) error {
	if err := WriteVarUint(buf, uint64(len(b))); err != nil {
		return err
	}
	_, err := buf.Write(b)
	return err
}

// WriteVarInt 编码一个有符号整数到 buf 中，使用变长编码
// 使用 ZigZag 编码将有符号整数转换为无符号整数，然后使用变长编码
func WriteVarInt(buf *bytes.Buffer, x int64) error {
	// ZigZag 编码：(x << 1) ^ (x >> 63)
	ux := uint64((x << 1) ^ (x >> 63))
	return WriteVarUint(buf, ux)
}

// WriteBool 编码一个布尔值到 buf 中，与 JS 的 readAny 兼容
func WriteBool(buf *bytes.Buffer, b bool) error {
	if b {
		return buf.WriteByte(120) // JS readAny uses 120 for true
	}
	return buf.WriteByte(121) // JS readAny uses 121 for false
}
