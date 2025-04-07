package bdd

import (
	"math/rand"
	"strconv"
)

// BooleanString 代表二进制字符串 "0" 或 "1"
type BooleanString string

const (
	BooleanZero BooleanString = "0"
	BooleanOne  BooleanString = "1"
)

// BooleanStringToBoolean 将二进制字符串转换为布尔值
func BooleanStringToBoolean(str BooleanString) bool {
	return str == BooleanOne
}

// BooleanToBooleanString 将布尔值转换为二进制字符串
func BooleanToBooleanString(b bool) BooleanString {
	if b {
		return BooleanOne
	}
	return BooleanZero
}

// OppositeBoolean 返回相反的二进制字符串
func OppositeBoolean(input BooleanString) BooleanString {
	if input == BooleanOne {
		return BooleanZero
	}
	return BooleanOne
}

// LastChar 返回字符串的最后一个字符
func LastChar(str string) string {
	if len(str) == 0 {
		return ""
	}
	return string(str[len(str)-1])
}

// DecimalToPaddedBinary 将十进制数转换为指定长度的二进制字符串
func DecimalToPaddedBinary(decimal int, padding int) string {
	binary := strconv.FormatInt(int64(decimal), 2)

	// 添加前导零以达到指定长度
	for len(binary) < padding {
		binary = "0" + binary
	}

	return binary
}

// BinaryToDecimal 将二进制字符串转换为十进制数
func BinaryToDecimal(binary string) int {
	decimal, _ := strconv.ParseInt(binary, 2, 64)
	return int(decimal)
}

// MinBinaryWithLength 返回指定长度的全0二进制字符串
func MinBinaryWithLength(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "0"
	}
	return result
}

// MaxBinaryWithLength 返回指定长度的全1二进制字符串
func MaxBinaryWithLength(length int) string {
	result := ""
	for i := 0; i < length; i++ {
		result += "1"
	}
	return result
}

// GetNextStateSet 获取二进制字符串的下一个状态
func GetNextStateSet(stateSet string) string {
	decimal := BinaryToDecimal(stateSet)
	increase := decimal + 1
	return DecimalToPaddedBinary(increase, len(stateSet))
}

// ShuffleArray 随机打乱数组
func ShuffleArray(arr interface{}) {
	switch v := arr.(type) {
	case []int:
		for i := len(v) - 1; i > 0; i-- {
			j := rand.Intn(i + 1)
			v[i], v[j] = v[j], v[i]
		}
	}
}
