package types

// 此文件导出所有类型和常量，方便从其他包引用

// 导出操作类型常量
const (
	// 操作类型常量导出
	Insert = OperationInsert
	Update = OperationUpdate
	Delete = OperationDelete
)

// 导出操作名称常量
const (
	// ActionName 常量
	DoNothing                             = ActionDoNothing
	InsertFirst                           = ActionInsertFirst
	InsertLast                            = ActionInsertLast
	RemoveFirstItem                       = ActionRemoveFirstItem
	RemoveLastItem                        = ActionRemoveLastItem
	RemoveFirstInsertLast                 = ActionRemoveFirstInsertLast
	RemoveLastInsertFirst                 = ActionRemoveLastInsertFirst
	RemoveFirstInsertFirst                = ActionRemoveFirstInsertFirst
	RemoveLastInsertLast                  = ActionRemoveLastInsertLast
	RemoveExisting                        = ActionRemoveExisting
	ReplaceExisting                       = ActionReplaceExisting
	AlwaysWrong                           = ActionAlwaysWrong
	InsertAtSortPosition                  = ActionInsertAtSortPosition
	RemoveExistingAndInsertAtSortPosition = ActionRemoveExistingAndInsertAtSortPosition
	RunFullQueryAgain                     = ActionRunFullQueryAgain
	UnknownAction                         = ActionUnknownAction
)

// 导出状态名称常量
const (
	// StateName 常量
	HasLimit             = StateHasLimit
	IsFindOne            = StateIsFindOne
	HasSkip              = StateHasSkip
	IsDelete             = StateIsDelete
	IsInsert             = StateIsInsert
	IsUpdate             = StateIsUpdate
	WasResultsEmpty      = StateWasResultsEmpty
	WasLimitReached      = StateWasLimitReached
	SortParamsChanged    = StateSortParamsChanged
	WasInResult          = StateWasInResult
	WasFirst             = StateWasFirst
	WasLast              = StateWasLast
	WasSortedBeforeFirst = StateWasSortedBeforeFirst
	WasSortedAfterLast   = StateWasSortedAfterLast
	IsSortedAfterLast    = StateIsSortedAfterLast
	IsSortedBeforeFirst  = StateIsSortedBeforeFirst
	WasMatching          = StateWasMatching
	DoesMatchNow         = StateDoesMatchNow
)
