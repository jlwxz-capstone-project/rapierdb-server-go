package eventreduce_test

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/eventreduce"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/state"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestGetStateIndex 测试获取状态索引的函数
func TestGetStateIndex(t *testing.T) {
	// 保存原始状态列表以便测试后恢复
	originalList := state.OrderedStateList
	defer func() { state.OrderedStateList = originalList }()

	// 设置测试用状态列表
	state.OrderedStateList = []types.StateName{
		"isDelete",
		"isInsert",
		"wasInResult",
		"doesMatchNow",
	}

	tests := []struct {
		name      string
		stateName string
		expected  int
	}{
		{"已知状态_isDelete", "isDelete", 0},
		{"已知状态_isInsert", "isInsert", 1},
		{"已知状态_wasInResult", "wasInResult", 2},
		{"已知状态_doesMatchNow", "doesMatchNow", 3},
		{"未知状态", "unknownState", -1},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := eventreduce.GetStateIndexForTest(tc.stateName)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGenerateAllStateSets 测试生成所有状态集合的函数
func TestGenerateAllStateSets(t *testing.T) {
	tests := []struct {
		name           string
		stateCount     int
		expectedLen    int
		expectContains []string
	}{
		{
			name:           "1位状态",
			stateCount:     1,
			expectedLen:    2,
			expectContains: []string{"0", "1"},
		},
		{
			name:           "2位状态",
			stateCount:     2,
			expectedLen:    4,
			expectContains: []string{"00", "01", "10", "11"},
		},
		{
			name:           "3位状态",
			stateCount:     3,
			expectedLen:    8,
			expectContains: []string{"000", "001", "010", "011", "100", "101", "110", "111"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := eventreduce.GenerateAllStateSetsForTest(tc.stateCount)

			// 检查生成的集合数量
			assert.Equal(t, tc.expectedLen, len(result))

			// 检查是否包含所有期望的状态集合
			for _, expected := range tc.expectContains {
				assert.Contains(t, result, expected)
			}
		})
	}
}

// 注意：要进行实际测试，需要在pkg/eventreduce/event_reduce.go中导出这些内部函数的测试版本
// 例如：
// func GetStateIndexForTest(stateName string) int {
//     return getStateIndex(stateName)
// }
//
// func GenerateAllStateSetsForTest(stateCount int) []string {
//     return generateAllStateSets(stateCount)
// }
