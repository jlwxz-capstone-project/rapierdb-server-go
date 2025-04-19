package eventreduce

// StateName 表示状态名称
type StateName string

const (
	// 状态类型常量
	StateHasLimit             StateName = "hasLimit"
	StateIsFindOne            StateName = "isFindOne"
	StateHasSkip              StateName = "hasSkip"
	StateIsDelete             StateName = "isDelete"
	StateIsInsert             StateName = "isInsert"
	StateIsUpdate             StateName = "isUpdate"
	StateWasResultsEmpty      StateName = "wasResultsEmpty"
	StateWasLimitReached      StateName = "wasLimitReached"
	StateSortParamsChanged    StateName = "sortParamsChanged"
	StateWasInResult          StateName = "wasInResult"
	StateWasFirst             StateName = "wasFirst"
	StateWasLast              StateName = "wasLast"
	StateWasSortedBeforeFirst StateName = "wasSortedBeforeFirst"
	StateWasSortedAfterLast   StateName = "wasSortedAfterLast"
	StateIsSortedAfterLast    StateName = "isSortedAfterLast"
	StateIsSortedBeforeFirst  StateName = "isSortedBeforeFirst"
	StateWasMatching          StateName = "wasMatching"
	StateDoesMatchNow         StateName = "doesMatchNow"
)

type StateSet = string

var OrderedStateNames = []StateName{
	StateHasLimit,
	StateIsFindOne,
	StateHasSkip,
	StateIsDelete,
	StateIsInsert,
	StateIsUpdate,
	StateWasResultsEmpty,
	StateWasLimitReached,
	StateSortParamsChanged,
	StateWasInResult,
	StateWasFirst,
	StateWasLast,
	StateWasSortedBeforeFirst,
	StateWasSortedAfterLast,
	StateIsSortedAfterLast,
	StateIsSortedBeforeFirst,
	StateWasMatching,
	StateDoesMatchNow,
}
