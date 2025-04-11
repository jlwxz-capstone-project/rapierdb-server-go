package state

import (
	"github.com/cockroachdb/errors"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// 导出所有状态解析函数
var (
	// 按性能开销排序的状态列表，最低开销在前
	OrderedStateList = []types.StateName{
		types.StateIsInsert,
		types.StateIsUpdate,
		types.StateIsDelete,
		types.StateHasLimit,
		types.StateIsFindOne,
		types.StateHasSkip,
		types.StateWasResultsEmpty,
		types.StateWasLimitReached,
		types.StateWasFirst,
		types.StateWasLast,
		types.StateSortParamsChanged,
		types.StateWasInResult,
		types.StateWasSortedBeforeFirst,
		types.StateWasSortedAfterLast,
		types.StateIsSortedBeforeFirst,
		types.StateIsSortedAfterLast,
		types.StateWasMatching,
		types.StateDoesMatchNow,
	}

	// 状态解析函数映射
	StateResolveFunctions = map[types.StateName]types.StateResolveFunction{
		types.StateIsInsert:             IsInsert,
		types.StateIsUpdate:             IsUpdate,
		types.StateIsDelete:             IsDelete,
		types.StateHasLimit:             HasLimit,
		types.StateIsFindOne:            IsFindOne,
		types.StateHasSkip:              HasSkip,
		types.StateWasResultsEmpty:      WasResultsEmpty,
		types.StateWasLimitReached:      WasLimitReached,
		types.StateWasFirst:             WasFirst,
		types.StateWasLast:              WasLast,
		types.StateSortParamsChanged:    SortParamsChanged,
		types.StateWasInResult:          WasInResult,
		types.StateWasSortedBeforeFirst: WasSortedBeforeFirst,
		types.StateWasSortedAfterLast:   WasSortedAfterLast,
		types.StateIsSortedBeforeFirst:  IsSortedBeforeFirst,
		types.StateIsSortedAfterLast:    IsSortedAfterLast,
		types.StateWasMatching:          WasMatching,
		types.StateDoesMatchNow:         DoesMatchNow,
	}

	// 按索引访问的状态解析函数
	StateResolveFunctionByIndex = map[int]types.StateResolveFunction{
		0:  IsInsert,
		1:  IsUpdate,
		2:  IsDelete,
		3:  HasLimit,
		4:  IsFindOne,
		5:  HasSkip,
		6:  WasResultsEmpty,
		7:  WasLimitReached,
		8:  WasFirst,
		9:  WasLast,
		10: SortParamsChanged,
		11: WasInResult,
		12: WasSortedBeforeFirst,
		13: WasSortedAfterLast,
		14: IsSortedBeforeFirst,
		15: IsSortedAfterLast,
		16: WasMatching,
		17: DoesMatchNow,
	}
)

// ResolveState 根据状态名称解析状态
func ResolveState(stateName types.StateName, input types.StateResolveFunctionInput) bool {
	fn, ok := StateResolveFunctions[stateName]
	if !ok {
		panic(errors.Newf("resolveState() has no function for %s", stateName))
	}
	return fn(input)
}

// GetStateSet 获取输入的状态集合
func GetStateSet(input types.StateResolveFunctionInput) types.StateSet {
	set := ""
	for _, name := range OrderedStateList {
		value := ResolveState(name, input)
		add := "0"
		if value {
			add = "1"
		}
		set += add
	}
	return types.StateSet(set)
}

// LogStateSet 打印状态集合
func LogStateSet(stateSet types.StateSet) {
	for i, state := range OrderedStateList {
		if i < len(stateSet) {
			println("state: " + string(state) + " : " + string(stateSet[i]))
		}
	}
}
