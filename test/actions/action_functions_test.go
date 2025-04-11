package actions_test

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/actions"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
	"github.com/stretchr/testify/assert"
)

// 创建测试用的输入数据
func createTestInput() *types.ActionFunctionInput {
	// 创建一个有3个文档的测试集
	previousResults := []map[string]interface{}{
		{
			"_id":    "doc1",
			"name":   "第一个文档",
			"active": true,
			"order":  1.0, // 使用float64类型
		},
		{
			"_id":    "doc2",
			"name":   "第二个文档",
			"active": true,
			"order":  2.0, // 使用float64类型
		},
		{
			"_id":    "doc3",
			"name":   "第三个文档",
			"active": false,
			"order":  3.0, // 使用float64类型
		},
	}

	// 创建键文档映射
	keyDocumentMap := make(map[string]map[string]interface{})
	for _, doc := range previousResults {
		id := doc["_id"].(string)
		keyDocumentMap[id] = doc
	}

	// 创建查询参数
	queryParams := types.QueryParams{
		PrimaryKey: "_id",
		SortComparator: func(a, b map[string]interface{}) int {
			// 支持float64类型的排序比较器
			var aOrder, bOrder float64

			// 安全的类型转换
			switch v := a["order"].(type) {
			case int:
				aOrder = float64(v)
			case float64:
				aOrder = v
			default:
				panic("order值类型不是int或float64")
			}

			switch v := b["order"].(type) {
			case int:
				bOrder = float64(v)
			case float64:
				bOrder = v
			default:
				panic("order值类型不是int或float64")
			}

			// 比较并返回结果
			if aOrder < bOrder {
				return -1
			} else if aOrder > bOrder {
				return 1
			}
			return 0
		},
	}

	// 创建一个新文档作为变更事件
	newDoc := map[string]interface{}{
		"_id":    "doc4",
		"name":   "新文档",
		"active": true,
		"order":  4.0, // 使用float64类型
	}
	changeEvent := types.NewInsertEvent("doc4", newDoc)

	// 返回测试输入
	return &types.ActionFunctionInput{
		QueryParams:     queryParams,
		ChangeEvent:     changeEvent,
		PreviousResults: previousResults,
		KeyDocumentMap:  keyDocumentMap,
	}
}

// 克隆测试输入，避免测试间相互影响
func cloneTestInput(input *types.ActionFunctionInput) *types.ActionFunctionInput {
	// 克隆previous results
	previousResults := make([]map[string]interface{}, len(input.PreviousResults))
	for i, doc := range input.PreviousResults {
		clonedDoc := make(map[string]interface{})
		for k, v := range doc {
			clonedDoc[k] = v
		}
		previousResults[i] = clonedDoc
	}

	// 克隆key document map
	keyDocumentMap := make(map[string]map[string]interface{})
	for k, doc := range input.KeyDocumentMap {
		clonedDoc := make(map[string]interface{})
		for docKey, docVal := range doc {
			clonedDoc[docKey] = docVal
		}
		keyDocumentMap[k] = clonedDoc
	}

	// 克隆changeEvent的Doc
	var clonedDoc map[string]interface{}
	if input.ChangeEvent.Doc != nil {
		clonedDoc = make(map[string]interface{})
		for k, v := range input.ChangeEvent.Doc {
			clonedDoc[k] = v
		}
	}

	// 克隆changeEvent的Previous
	var clonedPrevious map[string]interface{}
	if input.ChangeEvent.Previous != nil {
		clonedPrevious = make(map[string]interface{})
		for k, v := range input.ChangeEvent.Previous {
			clonedPrevious[k] = v
		}
	}

	// 创建克隆的changeEvent
	changeEvent := types.ChangeEvent{
		Operation: input.ChangeEvent.Operation,
		ID:        input.ChangeEvent.ID,
		Doc:       clonedDoc,
		Previous:  clonedPrevious,
	}

	// 返回克隆的输入
	return &types.ActionFunctionInput{
		QueryParams:     input.QueryParams,
		ChangeEvent:     changeEvent,
		PreviousResults: previousResults,
		KeyDocumentMap:  keyDocumentMap,
	}
}

// TestDoNothing 测试DoNothing函数
func TestDoNothing(t *testing.T) {
	input := createTestInput()
	originalLen := len(input.PreviousResults)
	originalMapLen := len(input.KeyDocumentMap)

	// 执行函数
	actions.DoNothing(input)

	// 验证结果
	assert.Equal(t, originalLen, len(input.PreviousResults), "结果集长度不应改变")
	assert.Equal(t, originalMapLen, len(input.KeyDocumentMap), "键文档映射长度不应改变")
}

// TestInsertFirst 测试InsertFirst函数
func TestInsertFirst(t *testing.T) {
	input := createTestInput()
	originalLen := len(input.PreviousResults)
	newDocID := input.ChangeEvent.ID

	// 执行函数
	actions.InsertFirst(input)

	// 验证结果
	assert.Equal(t, originalLen+1, len(input.PreviousResults), "结果集长度应增加1")
	assert.Equal(t, newDocID, input.PreviousResults[0]["_id"], "新文档应该在第一位")
	assert.Contains(t, input.KeyDocumentMap, newDocID, "键文档映射应包含新文档")
}

// TestInsertLast 测试InsertLast函数
func TestInsertLast(t *testing.T) {
	input := createTestInput()
	originalLen := len(input.PreviousResults)
	newDocID := input.ChangeEvent.ID

	// 执行函数
	actions.InsertLast(input)

	// 验证结果
	assert.Equal(t, originalLen+1, len(input.PreviousResults), "结果集长度应增加1")
	assert.Equal(t, newDocID, input.PreviousResults[len(input.PreviousResults)-1]["_id"], "新文档应该在最后一位")
	assert.Contains(t, input.KeyDocumentMap, newDocID, "键文档映射应包含新文档")
}

// TestRemoveExisting 测试RemoveExisting函数
func TestRemoveExisting(t *testing.T) {
	// 准备一个更新事件的输入
	baseInput := createTestInput()
	input := &types.ActionFunctionInput{
		QueryParams:     baseInput.QueryParams,
		ChangeEvent:     types.NewDeleteEvent("doc2", baseInput.KeyDocumentMap["doc2"]),
		PreviousResults: baseInput.PreviousResults,
		KeyDocumentMap:  baseInput.KeyDocumentMap,
	}

	originalLen := len(input.PreviousResults)
	docIDToRemove := "doc2"

	// 执行函数
	actions.RemoveExisting(input)

	// 验证结果
	assert.Equal(t, originalLen-1, len(input.PreviousResults), "结果集长度应减少1")
	for _, doc := range input.PreviousResults {
		assert.NotEqual(t, docIDToRemove, doc["_id"], "被删除的文档不应该在结果集中")
	}
	assert.NotContains(t, input.KeyDocumentMap, docIDToRemove, "键文档映射不应包含被删除的文档")
}

// TestReplaceExisting 测试ReplaceExisting函数
func TestReplaceExisting(t *testing.T) {
	// 准备一个更新事件的输入
	baseInput := createTestInput()
	updatedDoc := map[string]interface{}{
		"_id":    "doc2",
		"name":   "已更新的文档",
		"active": true,
		"order":  2.0, // 使用float64类型
	}

	input := &types.ActionFunctionInput{
		QueryParams:     baseInput.QueryParams,
		ChangeEvent:     types.NewUpdateEvent("doc2", updatedDoc, baseInput.KeyDocumentMap["doc2"]),
		PreviousResults: baseInput.PreviousResults,
		KeyDocumentMap:  baseInput.KeyDocumentMap,
	}

	originalLen := len(input.PreviousResults)
	docIDToUpdate := "doc2"

	// 执行函数
	actions.ReplaceExisting(input)

	// 验证结果
	assert.Equal(t, originalLen, len(input.PreviousResults), "结果集长度不应改变")

	// 查找更新后的文档
	var updatedDocInResults map[string]interface{}
	for _, doc := range input.PreviousResults {
		if doc["_id"] == docIDToUpdate {
			updatedDocInResults = doc
			break
		}
	}

	assert.NotNil(t, updatedDocInResults, "更新后的文档应该在结果集中")
	assert.Equal(t, "已更新的文档", updatedDocInResults["name"], "文档应该被更新")
	assert.Equal(t, updatedDoc, input.KeyDocumentMap[docIDToUpdate], "键文档映射中的文档应该被更新")
}
