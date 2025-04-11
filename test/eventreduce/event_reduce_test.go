package eventreduce_test

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/eventreduce"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/state"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// 创建测试用的修改事件
func createTestChangeEvent(operation types.WriteOperation, id string, doc, previous map[string]interface{}) types.ChangeEvent {
	switch operation {
	case types.OperationInsert:
		return types.NewInsertEvent(id, doc)
	case types.OperationUpdate:
		return types.NewUpdateEvent(id, doc, previous)
	case types.OperationDelete:
		return types.NewDeleteEvent(id, previous)
	default:
		return types.ChangeEvent{
			Operation: operation,
			ID:        id,
			Doc:       doc,
			Previous:  previous,
		}
	}
}

// 创建测试用的查询参数
func createTestQueryParams() types.QueryParams {
	queryMatcher := func(doc map[string]interface{}) bool {
		active, ok := doc["active"]
		return ok && active.(bool)
	}

	sortComparator := func(a, b map[string]interface{}) int {
		return 1 // 简单起见，总是返回1
	}

	limit := 10
	skip := 0

	return types.QueryParams{
		PrimaryKey:     "_id",
		SortFields:     []string{"name", "age"},
		Skip:           &skip,
		Limit:          &limit,
		QueryMatcher:   queryMatcher,
		SortComparator: sortComparator,
	}
}

// TestCalculateActionName 测试根据状态计算操作名称的功能
func TestCalculateActionName(t *testing.T) {
	// 保存原始状态列表以便测试后恢复
	originalList := state.OrderedStateList
	defer func() { state.OrderedStateList = originalList }()

	// 设置测试用状态列表 - 确保包含fillTruthTable函数中使用的所有状态
	state.OrderedStateList = []types.StateName{
		types.StateIsDelete,
		types.StateIsInsert,
		types.StateIsUpdate,
		types.StateWasInResult,
		types.StateDoesMatchNow,
		types.StateSortParamsChanged,
		types.StateIsSortedBeforeFirst,
		types.StateIsSortedAfterLast,
		types.StateWasMatching, // 虽然当前未使用，但为了完整性也包含
	}

	tests := []struct {
		name           string
		stateResolveFn map[types.StateName]func(types.StateResolveFunctionInput) bool
		expected       types.ActionName
	}{
		{
			name: "删除已存在的文档",
			stateResolveFn: map[types.StateName]func(types.StateResolveFunctionInput) bool{
				types.StateIsDelete:    func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateWasInResult: func(_ types.StateResolveFunctionInput) bool { return true },
			},
			expected: types.ActionRemoveExisting,
		},
		{
			name: "删除不存在的文档",
			stateResolveFn: map[types.StateName]func(types.StateResolveFunctionInput) bool{
				types.StateIsDelete:    func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateWasInResult: func(_ types.StateResolveFunctionInput) bool { return false },
			},
			expected: types.ActionDoNothing,
		},
		{
			name: "插入匹配的文档",
			stateResolveFn: map[types.StateName]func(types.StateResolveFunctionInput) bool{
				types.StateIsInsert:     func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateDoesMatchNow: func(_ types.StateResolveFunctionInput) bool { return true },
			},
			expected: types.ActionInsertAtSortPosition,
		},
		{
			name: "插入不匹配的文档",
			stateResolveFn: map[types.StateName]func(types.StateResolveFunctionInput) bool{
				types.StateIsInsert:     func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateDoesMatchNow: func(_ types.StateResolveFunctionInput) bool { return false },
			},
			expected: types.ActionDoNothing,
		},
		{
			name: "更新文档使其不再匹配",
			stateResolveFn: map[types.StateName]func(types.StateResolveFunctionInput) bool{
				types.StateIsUpdate:     func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateWasInResult:  func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateDoesMatchNow: func(_ types.StateResolveFunctionInput) bool { return false },
			},
			expected: types.ActionRemoveExisting,
		},
		{
			name: "更新文档保持匹配",
			stateResolveFn: map[types.StateName]func(types.StateResolveFunctionInput) bool{
				types.StateIsUpdate:     func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateWasInResult:  func(_ types.StateResolveFunctionInput) bool { return true },
				types.StateDoesMatchNow: func(_ types.StateResolveFunctionInput) bool { return true },
			},
			expected: types.ActionReplaceExisting,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 保存原始状态解析函数并在测试后恢复
			originalFns := state.StateResolveFunctions
			defer func() { state.StateResolveFunctions = originalFns }()

			// 设置测试用状态解析函数
			state.StateResolveFunctions = make(map[types.StateName]types.StateResolveFunction)

			// 默认情况下所有状态都返回false
			for _, name := range state.OrderedStateList {
				stateName := name
				state.StateResolveFunctions[stateName] = func(_ types.StateResolveFunctionInput) bool {
					return false
				}
			}

			// 设置测试用例中指定的状态解析函数
			for stateName, fn := range tc.stateResolveFn {
				state.StateResolveFunctions[stateName] = fn
			}

			// 创建测试输入
			input := types.StateResolveFunctionInput{
				QueryParams:     createTestQueryParams(),
				ChangeEvent:     createTestChangeEvent(types.OperationUpdate, "doc1", nil, nil),
				PreviousResults: []map[string]interface{}{},
				KeyDocumentMap:  make(map[string]map[string]interface{}),
			}

			// 调用被测试函数
			result := eventreduce.CalculateActionName(input)

			// 验证结果
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestRunAction 测试执行操作的功能
// 暂时跳过这个测试，因为我们需要模拟操作函数
func TestRunAction(t *testing.T) {
	t.Skip("需要更复杂的模拟环境，暂时跳过")

	// 创建测试数据和操作函数
	// 测试各种操作类型的结果
}
