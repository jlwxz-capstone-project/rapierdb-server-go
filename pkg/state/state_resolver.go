package state

import (
	"reflect"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// HasLimit 检查是否有limit参数
func HasLimit(input types.StateResolveFunctionInput) bool {
	return input.QueryParams.Limit != nil && *input.QueryParams.Limit > 0
}

// IsFindOne 检查是否是查找单条记录
func IsFindOne(input types.StateResolveFunctionInput) bool {
	return input.QueryParams.Limit != nil && *input.QueryParams.Limit == 1
}

// HasSkip 检查是否有skip参数
func HasSkip(input types.StateResolveFunctionInput) bool {
	return input.QueryParams.Skip != nil && *input.QueryParams.Skip > 0
}

// IsDelete 检查是否是删除操作
func IsDelete(input types.StateResolveFunctionInput) bool {
	return input.ChangeEvent.Operation == types.OperationDelete
}

// IsInsert 检查是否是插入操作
func IsInsert(input types.StateResolveFunctionInput) bool {
	return input.ChangeEvent.Operation == types.OperationInsert
}

// IsUpdate 检查是否是更新操作
func IsUpdate(input types.StateResolveFunctionInput) bool {
	return input.ChangeEvent.Operation == types.OperationUpdate
}

// WasLimitReached 检查是否已达到限制
func WasLimitReached(input types.StateResolveFunctionInput) bool {
	return HasLimit(input) && len(input.PreviousResults) >= *input.QueryParams.Limit
}

// SortParamsChanged 检查排序字段是否发生变化
func SortParamsChanged(input types.StateResolveFunctionInput) bool {
	sortFields := input.QueryParams.SortFields
	prev := input.ChangeEvent.Previous
	doc := input.ChangeEvent.Doc

	if doc == nil {
		return false
	}
	if prev == nil {
		return true
	}

	for _, field := range sortFields {
		beforeData := GetProperty(prev, field)
		afterData := GetProperty(doc, field)
		if !reflect.DeepEqual(beforeData, afterData) {
			return true
		}
	}
	return false
}

// WasInResult 检查之前是否在结果中
func WasInResult(input types.StateResolveFunctionInput) bool {
	id := input.ChangeEvent.ID
	if input.KeyDocumentMap != nil {
		_, has := input.KeyDocumentMap[id]
		return has
	} else {
		primary := input.QueryParams.PrimaryKey
		results := input.PreviousResults
		for _, item := range results {
			if item[primary] == id {
				return true
			}
		}
		return false
	}
}

// WasFirst 检查是否之前是第一条记录
func WasFirst(input types.StateResolveFunctionInput) bool {
	if len(input.PreviousResults) == 0 {
		return false
	}

	first := input.PreviousResults[0]
	return first[input.QueryParams.PrimaryKey] == input.ChangeEvent.ID
}

// WasLast 检查是否之前是最后一条记录
func WasLast(input types.StateResolveFunctionInput) bool {
	if len(input.PreviousResults) == 0 {
		return false
	}

	last := LastOfArray(input.PreviousResults)
	return last[input.QueryParams.PrimaryKey] == input.ChangeEvent.ID
}

// WasSortedBeforeFirst 检查之前是否排在第一条记录之前
func WasSortedBeforeFirst(input types.StateResolveFunctionInput) bool {
	prev := input.ChangeEvent.Previous
	if prev == nil {
		return false
	}

	if len(input.PreviousResults) == 0 {
		return false
	}

	first := input.PreviousResults[0]
	if first == nil {
		return false
	}

	// 如果变更的文档就是第一条，我们无法排序比较它们，因为可能会导致不确定的排序顺序
	// 因为两个文档可能相等，所以我们返回true
	if first[input.QueryParams.PrimaryKey] == input.ChangeEvent.ID {
		return true
	}

	comp := input.QueryParams.SortComparator(prev, first)
	return comp < 0
}

// WasSortedAfterLast 检查之前是否排在最后一条记录之后
func WasSortedAfterLast(input types.StateResolveFunctionInput) bool {
	prev := input.ChangeEvent.Previous
	if prev == nil {
		return false
	}

	if len(input.PreviousResults) == 0 {
		return false
	}

	last := LastOfArray(input.PreviousResults)
	if last == nil {
		return false
	}

	if last[input.QueryParams.PrimaryKey] == input.ChangeEvent.ID {
		return true
	}

	comp := input.QueryParams.SortComparator(prev, last)
	return comp > 0
}

// IsSortedBeforeFirst 检查当前是否排在第一条记录之前
func IsSortedBeforeFirst(input types.StateResolveFunctionInput) bool {
	doc := input.ChangeEvent.Doc
	if doc == nil {
		return false
	}

	if len(input.PreviousResults) == 0 {
		return false
	}

	first := input.PreviousResults[0]
	if first == nil {
		return false
	}

	if first[input.QueryParams.PrimaryKey] == input.ChangeEvent.ID {
		return true
	}

	comp := input.QueryParams.SortComparator(doc, first)
	return comp < 0
}

// IsSortedAfterLast 检查当前是否排在最后一条记录之后
func IsSortedAfterLast(input types.StateResolveFunctionInput) bool {
	doc := input.ChangeEvent.Doc
	if doc == nil {
		return false
	}

	if len(input.PreviousResults) == 0 {
		return false
	}

	last := LastOfArray(input.PreviousResults)
	if last == nil {
		return false
	}

	if last[input.QueryParams.PrimaryKey] == input.ChangeEvent.ID {
		return true
	}

	comp := input.QueryParams.SortComparator(doc, last)
	return comp > 0
}

// WasMatching 检查之前是否匹配查询条件
func WasMatching(input types.StateResolveFunctionInput) bool {
	prev := input.ChangeEvent.Previous
	if prev == nil {
		return false
	}
	return input.QueryParams.QueryMatcher(prev)
}

// DoesMatchNow 检查当前是否匹配查询条件
func DoesMatchNow(input types.StateResolveFunctionInput) bool {
	doc := input.ChangeEvent.Doc
	if doc == nil {
		return false
	}
	return input.QueryParams.QueryMatcher(doc)
}

// WasResultsEmpty 检查之前的结果是否为空
func WasResultsEmpty(input types.StateResolveFunctionInput) bool {
	return len(input.PreviousResults) == 0
}

// GetProperty 从对象中获取属性值
func GetProperty(obj map[string]interface{}, path string) interface{} {
	if obj == nil {
		return nil
	}

	return obj[path]
}

// LastOfArray 获取数组的最后一个元素
func LastOfArray(arr []map[string]interface{}) map[string]interface{} {
	if len(arr) == 0 {
		return nil
	}
	return arr[len(arr)-1]
}
