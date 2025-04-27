package util

import (
	"bytes"
	"encoding/binary"
	"errors"
)

var ErrVarintOverflow = errors.New("varint overflow")

// ReadUint8 从 buf 中解码一个无符号 8 位整数
func ReadUint8(buf *bytes.Buffer) (uint8, error) {
	return buf.ReadByte()
}

// ReadVarUint 从 buf 中解码一个无符号整数，
// 使用变长编码，每个字节的高位表示是否继续，低位表示数值
func ReadVarUint(buf *bytes.Buffer) (uint64, error) {
	var x uint64
	var s uint
	for i := 0; ; i++ {
		b, err := buf.ReadByte()
		if err != nil {
			return 0, err
		}
		if b < 0x80 {
			if i > 9 || i == 9 && b > 1 {
				return 0, ErrVarintOverflow
			}
			return x | uint64(b)<<s, nil
		}
		x |= uint64(b&0x7f) << s
		s += 7
	}
}

// ReadVarString 从 buf 中解码一个字符串
// 先读取字符串的长度，然后读取字符串内容
func ReadVarString(buf *bytes.Buffer) (string, error) {
	length, err := ReadVarUint(buf)
	if err != nil {
		return "", err
	}
	bytes := make([]byte, length)
	_, err = buf.Read(bytes)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ReadUint32 从 buf 中解码一个无符号 32 位整数，使用小端编码，大小为 4 字节
func ReadUint32(buf *bytes.Buffer) (uint32, error) {
	var n uint32
	err := binary.Read(buf, binary.LittleEndian, &n)
	return n, err
}

// ReadUint64 从 buf 中解码一个无符号 64 位整数，使用小端编码，大小为 8 字节
func ReadUint64(buf *bytes.Buffer) (uint64, error) {
	var n uint64
	err := binary.Read(buf, binary.LittleEndian, &n)
	return n, err
}

// ReadInt8 从 buf 中解码一个有符号 8 位整数
func ReadInt8(buf *bytes.Buffer) (int8, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return 0, err
	}
	return int8(b), nil
}

// ReadInt16 从 buf 中解码一个有符号 16 位整数，使用小端编码
func ReadInt16(buf *bytes.Buffer) (int16, error) {
	var n int16
	err := binary.Read(buf, binary.LittleEndian, &n)
	return n, err
}

// ReadInt32 从 buf 中解码一个有符号 32 位整数，使用小端编码
func ReadInt32(buf *bytes.Buffer) (int32, error) {
	var n int32
	err := binary.Read(buf, binary.LittleEndian, &n)
	return n, err
}

// ReadInt64 从 buf 中解码一个有符号 64 位整数，使用小端编码
func ReadInt64(buf *bytes.Buffer) (int64, error) {
	var n int64
	err := binary.Read(buf, binary.LittleEndian, &n)
	return n, err
}

// ReadFloat32 从 buf 中解码一个 32 位浮点数，使用小端编码
func ReadFloat32(buf *bytes.Buffer) (float32, error) {
	var n float32
	err := binary.Read(buf, binary.LittleEndian, &n)
	return n, err
}

// ReadFloat64 从 buf 中解码一个 64 位浮点数，使用小端编码
func ReadFloat64(buf *bytes.Buffer) (float64, error) {
	var n float64
	err := binary.Read(buf, binary.LittleEndian, &n)
	return n, err
}

// ReadBool 从 buf 中解码一个布尔值
func ReadBool(buf *bytes.Buffer) (bool, error) {
	b, err := buf.ReadByte()
	if err != nil {
		return false, err
	}
	return b != 0, nil
}

// ReadVarByteArray 从 buf 中解码一个字节数组
// 先读取数组的长度，然后读取数组内容
func ReadVarByteArray(buf *bytes.Buffer) ([]byte, error) {
	length, err := ReadVarUint(buf)
	if err != nil {
		return nil, err
	}
	bytes := make([]byte, length)
	_, err = buf.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// ReadVarInt 从 buf 中解码一个有符号整数，使用变长编码
// 使用变长编码读取 ZigZag 编码的无符号整数，然后转换回有符号整数
func ReadVarInt(buf *bytes.Buffer) (int64, error) {
	ux, err := ReadVarUint(buf)
	if err != nil {
		return 0, err
	}
	// 解码 ZigZag：(ux >> 1) ^ -(ux & 1)
	x := int64(ux >> 1)
	if ux&1 != 0 {
		x = ^x
	}
	return x, nil
}
