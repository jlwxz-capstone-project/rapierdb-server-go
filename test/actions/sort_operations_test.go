package actions_test

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/actions"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// TestPushAtSortPosition 测试PushAtSortPosition函数
func TestPushAtSortPosition(t *testing.T) {
	// 准备测试数据
	arr := []map[string]interface{}{
		{"id": 1, "value": 10.0},
		{"id": 2, "value": 30.0},
		{"id": 3, "value": 50.0},
	}

	// 定义排序比较器
	comparator := func(a, b map[string]interface{}) int {
		// 安全的类型处理
		var aVal, bVal float64

		switch v := a["value"].(type) {
		case int:
			aVal = float64(v)
		case float64:
			aVal = v
		default:
			panic("value值类型不是int或float64")
		}

		switch v := b["value"].(type) {
		case int:
			bVal = float64(v)
		case float64:
			bVal = v
		default:
			panic("value值类型不是int或float64")
		}

		if aVal < bVal {
			return -1
		} else if aVal > bVal {
			return 1
		}
		return 0
	}

	tests := []struct {
		name     string
		doc      map[string]interface{}
		expected int // 期望插入位置的索引
	}{
		{
			name:     "插入到开头",
			doc:      map[string]interface{}{"id": 0, "value": 5.0},
			expected: 0,
		},
		{
			name:     "插入到中间",
			doc:      map[string]interface{}{"id": 2.5, "value": 40.0},
			expected: 2,
		},
		{
			name:     "插入到末尾",
			doc:      map[string]interface{}{"id": 4, "value": 60.0},
			expected: 3,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// 复制原始数组
			original := make([]map[string]interface{}, len(arr))
			copy(original, arr)

			// 执行函数
			result := actions.PushAtSortPosition(original, tc.doc, comparator)

			// 验证结果
			assert.Equal(t, len(original)+1, len(result), "结果数组长度应比原始数组长度多1")
			assert.Equal(t, tc.doc, result[tc.expected], "文档应该在预期的位置")
		})
	}
}

// TestInsertAtSortPosition 测试InsertAtSortPosition函数
func TestInsertAtSortPosition(t *testing.T) {
	// 创建测试输入
	input := createTestInput()

	// 修改新文档的order值，使其应该插入到中间位置
	input.ChangeEvent.Doc["order"] = 2.5

	originalLen := len(input.PreviousResults)
	newDocID := input.ChangeEvent.ID

	// 执行函数
	actions.InsertAtSortPosition(input)

	// 验证结果
	assert.Equal(t, originalLen+1, len(input.PreviousResults), "结果集长度应增加1")

	// 验证文档位置是否正确（应该在索引2的位置，即第3个元素）
	assert.Equal(t, newDocID, input.PreviousResults[2]["_id"], "新文档应该在排序后的正确位置")
	assert.Contains(t, input.KeyDocumentMap, newDocID, "键文档映射应包含新文档")
}

// TestRemoveExistingAndInsertAtSortPosition 测试RemoveExistingAndInsertAtSortPosition函数
func TestRemoveExistingAndInsertAtSortPosition(t *testing.T) {
	baseInput := createTestInput()

	// 创建一个已存在但顺序改变的文档
	updatedDoc := map[string]interface{}{
		"_id":    "doc1", // 原来在第一位
		"name":   "位置变更的文档",
		"active": true,
		"order":  3.5, // 新的排序值，应该在第3个位置之后
	}

	input := &types.ActionFunctionInput{
		QueryParams:     baseInput.QueryParams,
		ChangeEvent:     types.NewUpdateEvent("doc1", updatedDoc, baseInput.KeyDocumentMap["doc1"]),
		PreviousResults: baseInput.PreviousResults,
		KeyDocumentMap:  baseInput.KeyDocumentMap,
	}

	originalLen := len(input.PreviousResults)
	docID := "doc1"

	// 执行函数
	actions.RemoveExistingAndInsertAtSortPosition(input)

	// 验证结果
	assert.Equal(t, originalLen, len(input.PreviousResults), "结果集长度应保持不变")

	// 文档原来在第一位，现在应该在第三位
	assert.Equal(t, docID, input.PreviousResults[2]["_id"], "更新后的文档应该在新的排序位置")
	assert.Contains(t, input.KeyDocumentMap, docID, "键文档映射应包含更新后的文档")
	assert.Equal(t, updatedDoc, input.KeyDocumentMap[docID], "键文档映射中的文档应该被更新")
}

// TestInsertAtSortPositionWithEmptyResults 测试空结果集的情况
func TestInsertAtSortPositionWithEmptyResults(t *testing.T) {
	input := createTestInput()

	// 清空结果集
	input.PreviousResults = []map[string]interface{}{}
	input.KeyDocumentMap = make(map[string]map[string]interface{})

	newDocID := input.ChangeEvent.ID

	// 执行函数
	actions.InsertAtSortPosition(input)

	// 验证结果
	assert.Equal(t, 1, len(input.PreviousResults), "结果集长度应为1")
	assert.Equal(t, newDocID, input.PreviousResults[0]["_id"], "新文档应该是唯一的元素")
	assert.Contains(t, input.KeyDocumentMap, newDocID, "键文档映射应包含新文档")
}
