package actions

import (
	"errors"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// DoNothing 什么也不做
func DoNothing(input *types.ActionFunctionInput) {
	// 什么也不做
}

// InsertFirst 在结果集开头插入文档
func InsertFirst(input *types.ActionFunctionInput) {
	// 在数组开头插入文档
	input.PreviousResults = append([]map[string]interface{}{input.ChangeEvent.Doc}, input.PreviousResults...)

	// 如果存在键文档映射，则更新
	if input.KeyDocumentMap != nil {
		input.KeyDocumentMap[input.ChangeEvent.ID] = input.ChangeEvent.Doc
	}
}

// InsertLast 在结果集末尾插入文档
func InsertLast(input *types.ActionFunctionInput) {
	// 在数组末尾插入文档
	input.PreviousResults = append(input.PreviousResults, input.ChangeEvent.Doc)

	// 如果存在键文档映射，则更新
	if input.KeyDocumentMap != nil {
		input.KeyDocumentMap[input.ChangeEvent.ID] = input.ChangeEvent.Doc
	}
}

// RemoveFirstItem 移除结果集中的第一项
func RemoveFirstItem(input *types.ActionFunctionInput) {
	if len(input.PreviousResults) == 0 {
		return
	}

	// 获取第一个元素
	first := input.PreviousResults[0]

	// 从数组中移除第一项
	input.PreviousResults = input.PreviousResults[1:]

	// 如果存在键文档映射，则更新
	if input.KeyDocumentMap != nil {
		delete(input.KeyDocumentMap, first[input.QueryParams.PrimaryKey].(string))
	}
}

// RemoveLastItem 移除结果集中的最后一项
func RemoveLastItem(input *types.ActionFunctionInput) {
	if len(input.PreviousResults) == 0 {
		return
	}

	// 获取最后一个元素
	last := input.PreviousResults[len(input.PreviousResults)-1]

	// 从数组中移除最后一项
	input.PreviousResults = input.PreviousResults[:len(input.PreviousResults)-1]

	// 如果存在键文档映射，则更新
	if input.KeyDocumentMap != nil {
		delete(input.KeyDocumentMap, last[input.QueryParams.PrimaryKey].(string))
	}
}

// RemoveFirstInsertLast 移除第一项并在末尾插入
func RemoveFirstInsertLast(input *types.ActionFunctionInput) {
	RemoveFirstItem(input)
	InsertLast(input)
}

// RemoveLastInsertFirst 移除最后一项并在开头插入
func RemoveLastInsertFirst(input *types.ActionFunctionInput) {
	RemoveLastItem(input)
	InsertFirst(input)
}

// RemoveFirstInsertFirst 移除第一项并在开头插入
func RemoveFirstInsertFirst(input *types.ActionFunctionInput) {
	RemoveFirstItem(input)
	InsertFirst(input)
}

// RemoveLastInsertLast 移除最后一项并在末尾插入
func RemoveLastInsertLast(input *types.ActionFunctionInput) {
	RemoveLastItem(input)
	InsertLast(input)
}

// RemoveExisting 移除已存在的项
func RemoveExisting(input *types.ActionFunctionInput) {
	// 如果存在键文档映射，则直接删除
	if input.KeyDocumentMap != nil {
		delete(input.KeyDocumentMap, input.ChangeEvent.ID)
	}

	// 查找文档的索引
	primary := input.QueryParams.PrimaryKey
	results := input.PreviousResults

	for i, item := range results {
		if item[primary] == input.ChangeEvent.ID {
			// 移除该项
			input.PreviousResults = append(results[:i], results[i+1:]...)
			break
		}
	}
}

// ReplaceExisting 替换已存在的项
func ReplaceExisting(input *types.ActionFunctionInput) {
	doc := input.ChangeEvent.Doc
	primary := input.QueryParams.PrimaryKey
	results := input.PreviousResults

	for i, item := range results {
		if item[primary] == input.ChangeEvent.ID {
			// 替换该项
			input.PreviousResults[i] = doc

			// 更新键文档映射
			if input.KeyDocumentMap != nil {
				input.KeyDocumentMap[input.ChangeEvent.ID] = doc
			}
			break
		}
	}
}

// AlwaysWrong 总是返回错误的结果
// 该函数必须在后续步骤中被优化掉，否则就是有问题的
func AlwaysWrong(input *types.ActionFunctionInput) {
	wrongHuman := map[string]interface{}{
		"_id": "wrongHuman" + time.Now().Format("20060102150405.000"),
	}

	// 清空数组
	input.PreviousResults = []map[string]interface{}{wrongHuman}

	// 更新键文档映射
	if input.KeyDocumentMap != nil {
		// 清空映射
		for k := range input.KeyDocumentMap {
			delete(input.KeyDocumentMap, k)
		}

		input.KeyDocumentMap[wrongHuman["_id"].(string)] = wrongHuman
	}
}

// PushAtSortPosition 在排序位置插入元素
func PushAtSortPosition(arr []map[string]interface{}, doc map[string]interface{}, comparator types.DeterministicSortComparator) []map[string]interface{} {
	// 如果数组为空，直接添加并返回
	if len(arr) == 0 {
		return append(arr, doc)
	}

	// 二分查找插入位置
	left := 0
	right := len(arr) - 1
	insertPos := len(arr) // 默认在末尾插入

	for left <= right {
		mid := (left + right) / 2
		comp := comparator(doc, arr[mid])

		if comp < 0 {
			// doc 应该在 mid 之前
			insertPos = mid
			right = mid - 1
		} else {
			// doc 应该在 mid 之后
			left = mid + 1
		}
	}

	// 在插入位置处插入元素
	if insertPos == len(arr) {
		return append(arr, doc)
	}

	result := make([]map[string]interface{}, 0, len(arr)+1)
	result = append(result, arr[:insertPos]...)
	result = append(result, doc)
	result = append(result, arr[insertPos:]...)

	return result
}

// InsertAtSortPosition 在排序位置插入文档
func InsertAtSortPosition(input *types.ActionFunctionInput) {
	docID := input.ChangeEvent.ID
	doc := input.ChangeEvent.Doc

	if input.KeyDocumentMap != nil {
		// 如果文档已经在结果中，不能再次添加，因为这会导致不确定的排序
		if _, exists := input.KeyDocumentMap[docID]; exists {
			return
		}

		// 更新键文档映射
		input.KeyDocumentMap[docID] = doc
	} else {
		// 检查文档是否已经在结果中
		primary := input.QueryParams.PrimaryKey
		for _, item := range input.PreviousResults {
			if item[primary] == docID {
				// 文档已经在结果中，不能再次添加
				return
			}
		}
	}

	// 在排序位置插入文档
	input.PreviousResults = PushAtSortPosition(input.PreviousResults, doc, input.QueryParams.SortComparator)
}

// RemoveExistingAndInsertAtSortPosition 移除已存在的项并在排序位置插入
func RemoveExistingAndInsertAtSortPosition(input *types.ActionFunctionInput) {
	RemoveExisting(input)
	InsertAtSortPosition(input)
}

// RunFullQueryAgain 重新运行完整查询
func RunFullQueryAgain(input *types.ActionFunctionInput) {
	panic(errors.New("Action runFullQueryAgain must be implemented by yourself"))
}

// UnknownAction 未知操作，不应该被调用
func UnknownAction(input *types.ActionFunctionInput) {
	panic(errors.New("Action unknownAction should never be called"))
}
