package state_test

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/state"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// 创建测试用的输入数据
func createTestInput(op types.WriteOperation) types.StateResolveFunctionInput {
	// 创建基本的查询参数
	queryParams := types.QueryParams{
		PrimaryKey: "_id",
		QueryMatcher: func(doc map[string]interface{}) bool {
			active, exists := doc["active"]
			return exists && active.(bool)
		},
	}

	// 创建不同操作类型的事件
	var event types.ChangeEvent
	switch op {
	case types.OperationInsert:
		event = types.NewInsertEvent("123", map[string]interface{}{
			"_id":    "123",
			"active": true,
			"name":   "Test",
		})
	case types.OperationUpdate:
		event = types.NewUpdateEvent("123",
			map[string]interface{}{
				"_id":    "123",
				"active": true,
				"name":   "Updated",
			},
			map[string]interface{}{
				"_id":    "123",
				"active": true,
				"name":   "Original",
			})
	case types.OperationDelete:
		event = types.NewDeleteEvent("123", map[string]interface{}{
			"_id":    "123",
			"active": true,
			"name":   "ToDelete",
		})
	}

	// 创建测试输入
	return types.StateResolveFunctionInput{
		QueryParams:     queryParams,
		ChangeEvent:     event,
		PreviousResults: []map[string]interface{}{},
		KeyDocumentMap:  map[string]map[string]interface{}{},
	}
}

// TestStateResolveFunctions 测试所有状态解析函数
func TestStateResolveFunctions(t *testing.T) {
	tests := []struct {
		name      string
		stateName types.StateName
		input     types.StateResolveFunctionInput
		expected  bool
	}{
		{
			name:      "IsInsert - 插入操作",
			stateName: types.StateIsInsert,
			input:     createTestInput(types.OperationInsert),
			expected:  true,
		},
		{
			name:      "IsInsert - 非插入操作",
			stateName: types.StateIsInsert,
			input:     createTestInput(types.OperationUpdate),
			expected:  false,
		},
		{
			name:      "IsUpdate - 更新操作",
			stateName: types.StateIsUpdate,
			input:     createTestInput(types.OperationUpdate),
			expected:  true,
		},
		{
			name:      "IsUpdate - 非更新操作",
			stateName: types.StateIsUpdate,
			input:     createTestInput(types.OperationInsert),
			expected:  false,
		},
		{
			name:      "IsDelete - 删除操作",
			stateName: types.StateIsDelete,
			input:     createTestInput(types.OperationDelete),
			expected:  true,
		},
		{
			name:      "IsDelete - 非删除操作",
			stateName: types.StateIsDelete,
			input:     createTestInput(types.OperationUpdate),
			expected:  false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := state.ResolveState(tc.stateName, tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestGetStateSet 测试状态集合生成
func TestGetStateSet(t *testing.T) {
	// 测试插入操作的状态集合
	input := createTestInput(types.OperationInsert)
	stateSet := state.GetStateSet(input)

	// 确保结果是一个字符串，长度等于状态列表长度
	assert.Equal(t, len(state.OrderedStateList), len(stateSet))

	// 验证插入操作的状态 - IsInsert应该为1
	isInsertIdx := -1
	for i, name := range state.OrderedStateList {
		if name == types.StateIsInsert {
			isInsertIdx = i
			break
		}
	}

	if isInsertIdx >= 0 {
		// 将字符转换为字符串后比较
		assert.Equal(t, "1", string(stateSet[isInsertIdx]), "IsInsert状态应该为1")
	}

	// 验证插入操作的状态 - IsDelete应该为0
	isDeleteIdx := -1
	for i, name := range state.OrderedStateList {
		if name == types.StateIsDelete {
			isDeleteIdx = i
			break
		}
	}

	if isDeleteIdx >= 0 {
		// 将字符转换为字符串后比较
		assert.Equal(t, "0", string(stateSet[isDeleteIdx]), "IsDelete状态应该为0")
	}
}
