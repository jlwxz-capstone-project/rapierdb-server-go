package bdd

import (
	"math/rand"
	"time"
)

const Unknown = 42 // 未知值

// TruthTable 真值表类型，将二进制状态字符串映射到值
type TruthTable map[string]int

// FillTruthTable 用给定值填充真值表中缺失的行
func FillTruthTable(truthTable TruthTable, inputLength int, value int) {
	endInput := MaxBinaryWithLength(inputLength)
	currentInput := MinBinaryWithLength(inputLength)

	done := false
	for !done {
		if _, exists := truthTable[currentInput]; !exists {
			truthTable[currentInput] = value
		}

		if currentInput == endInput {
			done = true
		} else {
			currentInput = GetNextStateSet(currentInput)
		}
	}
}

// NewExampleTruthTable 创建示例真值表
func NewExampleTruthTable(stateLength int) TruthTable {
	lastID := 0
	ret := make(TruthTable)
	maxBin := MaxBinaryWithLength(stateLength)
	maxDecimal := BinaryToDecimal(maxBin)

	end := maxDecimal
	start := 0
	for start <= end {
		ret[DecimalToPaddedBinary(start, stateLength)] = lastID
		lastID++
		start++
	}

	return ret
}

// NewAllEqualTable 创建所有元素相等的真值表
func NewAllEqualTable(stateLength int) TruthTable {
	if stateLength <= 0 {
		stateLength = 3
	}

	table := NewExampleTruthTable(stateLength)
	for k := range table {
		table[k] = 1
	}
	return table
}

// NewRandomTable 创建随机真值表
func NewRandomTable(stateLength int) TruthTable {
	if stateLength <= 0 {
		stateLength = 3
	}

	rand.Seed(time.Now().UnixNano())
	table := NewExampleTruthTable(stateLength)
	for k := range table {
		// "2"出现的几率比"1"高
		val := 2
		if rand.Float32() < 0.25 { // 25%的几率是1
			val = 1
		}
		table[k] = val
	}
	return table
}

// NewRandomUnknownTable 创建带未知值的随机真值表
func NewRandomUnknownTable(stateLength int) TruthTable {
	if stateLength <= 0 {
		stateLength = 3
	}

	rand.Seed(time.Now().UnixNano())
	table := NewExampleTruthTable(stateLength)
	for k := range table {
		if rand.Float32() < 0.5 { // 50%的几率是未知值
			table[k] = Unknown
		}
	}
	return table
}
