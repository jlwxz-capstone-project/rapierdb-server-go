package bdd

import (
	"math/rand"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/orderedmap"
)

const Unknown = 42 // 未知值

// TruthTable 真值表类型，将二进制状态字符串映射到值
type TruthTable = *orderedmap.OrderedMap[string, int]

// FillTruthTable 用给定值填充真值表中缺失的行
func FillTruthTable(truthTable TruthTable, inputLength int, value int) {
	endInput := MaxBinaryWithLength(inputLength)
	currentInput := MinBinaryWithLength(inputLength)

	done := false
	for !done {
		if _, exists := truthTable.Get(currentInput); !exists {
			truthTable.Set(currentInput, value)
		}

		if currentInput == endInput {
			done = true
		} else {
			currentInput = GetNextStateSet(currentInput)
		}
	}
}

// NewEmptyTruthTable 创建空真值表
func NewEmptyTruthTable() TruthTable {
	return orderedmap.NewOrderedMap[string, int]()
}

// NewExampleTruthTable 创建示例真值表
func NewExampleTruthTable(stateLength int) TruthTable {
	lastID := 0
	ret := orderedmap.NewOrderedMap[string, int]()
	maxBin := MaxBinaryWithLength(stateLength)
	maxDecimal := BinaryToDecimal(maxBin)

	end := maxDecimal
	start := 0
	for start <= end {
		ret.Set(DecimalToPaddedBinary(start, stateLength), lastID)
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
	for k := range table.IterKeys() {
		table.Set(k, 1)
	}
	return table
}

func randomBoolean() bool {
	return rand.Float32() < 0.5
}

// NewRandomTable 创建随机真值表
func NewRandomTable(stateLength int) TruthTable {
	rand.Seed(time.Now().UnixNano())
	table := NewExampleTruthTable(stateLength)
	for k := range table.IterKeys() {
		var val int
		if randomBoolean() && randomBoolean() {
			val = 1
		} else {
			val = 2
		}
		table.Set(k, val)
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
	for k := range table.IterKeys() {
		if rand.Float32() < 0.5 { // 50%的几率是未知值
			table.Set(k, Unknown)
		}
	}
	return table
}
