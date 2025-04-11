package state

import "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"

// 导出所有函数、类型和常量
// 这个文件存在的目的是为了声明所有导出符号，方便导入时使用

// 导出状态解析函数
var (
	// 状态解析函数
	HasLimitFunc             types.StateResolveFunction = HasLimit
	IsFindOneFunc            types.StateResolveFunction = IsFindOne
	HasSkipFunc              types.StateResolveFunction = HasSkip
	IsDeleteFunc             types.StateResolveFunction = IsDelete
	IsInsertFunc             types.StateResolveFunction = IsInsert
	IsUpdateFunc             types.StateResolveFunction = IsUpdate
	WasLimitReachedFunc      types.StateResolveFunction = WasLimitReached
	SortParamsChangedFunc    types.StateResolveFunction = SortParamsChanged
	WasInResultFunc          types.StateResolveFunction = WasInResult
	WasFirstFunc             types.StateResolveFunction = WasFirst
	WasLastFunc              types.StateResolveFunction = WasLast
	WasSortedBeforeFirstFunc types.StateResolveFunction = WasSortedBeforeFirst
	WasSortedAfterLastFunc   types.StateResolveFunction = WasSortedAfterLast
	IsSortedBeforeFirstFunc  types.StateResolveFunction = IsSortedBeforeFirst
	IsSortedAfterLastFunc    types.StateResolveFunction = IsSortedAfterLast
	WasMatchingFunc          types.StateResolveFunction = WasMatching
	DoesMatchNowFunc         types.StateResolveFunction = DoesMatchNow
	WasResultsEmptyFunc      types.StateResolveFunction = WasResultsEmpty

	// 工具函数
	GetPropertyFunc = GetProperty
	LastOfArrayFunc = LastOfArray

	// 状态管理函数
	ResolveStateFunc = ResolveState
	GetStateSetFunc  = GetStateSet
	LogStateSetFunc  = LogStateSet
)
