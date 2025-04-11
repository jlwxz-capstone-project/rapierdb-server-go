package types

// ActionName 表示操作名称
type ActionName string

const (
	// 操作类型常量
	ActionDoNothing                             ActionName = "doNothing"
	ActionInsertFirst                           ActionName = "insertFirst"
	ActionInsertLast                            ActionName = "insertLast"
	ActionRemoveFirstItem                       ActionName = "removeFirstItem"
	ActionRemoveLastItem                        ActionName = "removeLastItem"
	ActionRemoveFirstInsertLast                 ActionName = "removeFirstInsertLast"
	ActionRemoveLastInsertFirst                 ActionName = "removeLastInsertFirst"
	ActionRemoveFirstInsertFirst                ActionName = "removeFirstInsertFirst"
	ActionRemoveLastInsertLast                  ActionName = "removeLastInsertLast"
	ActionRemoveExisting                        ActionName = "removeExisting"
	ActionReplaceExisting                       ActionName = "replaceExisting"
	ActionAlwaysWrong                           ActionName = "alwaysWrong" // 这应该在后续步骤中被优化掉
	ActionInsertAtSortPosition                  ActionName = "insertAtSortPosition"
	ActionRemoveExistingAndInsertAtSortPosition ActionName = "removeExistingAndInsertAtSortPosition"
	ActionRunFullQueryAgain                     ActionName = "runFullQueryAgain"
	ActionUnknownAction                         ActionName = "unknownAction" // 如果状态从未达到，我们无法知道正确的操作
)

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

// ResultKeyDocumentMap 表示键到文档的映射
type ResultKeyDocumentMap map[string]map[string]interface{}

// QueryMatcher 表示查询匹配器函数
type QueryMatcher func(doc map[string]interface{}) bool

// DeterministicSortComparator 表示确定性排序比较器
// 为了使排序具有确定性，我们不返回0，只返回1或-1
// 这确保我们总是得到相同的输出数组，无论输入数组的预排序如何
type DeterministicSortComparator func(a, b map[string]interface{}) int

// StateSet 表示状态集，是有序状态列表的二进制表示
// 例如 '010110110111...'，其中第一个'0'表示第一个状态(hasLimit)为false
type StateSet string

// StateSetToActionMap 表示状态集到操作的映射
// 不在映射中的状态集具有'runFullQueryAgain'值
type StateSetToActionMap map[StateSet]ActionName

// QueryParams 表示查询参数
type QueryParams struct {
	PrimaryKey     string                      `json:"primaryKey"`
	SortFields     []string                    `json:"sortFields"`
	Skip           *int                        `json:"skip,omitempty"`
	Limit          *int                        `json:"limit,omitempty"`
	QueryMatcher   QueryMatcher                `json:"queryMatcher"`
	SortComparator DeterministicSortComparator `json:"sortComparator"`
}

// StateResolveFunctionInput 表示状态解析函数的输入
type StateResolveFunctionInput struct {
	QueryParams     QueryParams                       `json:"queryParams"`
	ChangeEvent     ChangeEvent                       `json:"changeEvent"`
	PreviousResults []map[string]interface{}          `json:"previousResults"`
	KeyDocumentMap  map[string]map[string]interface{} `json:"keyDocumentMap,omitempty"`
}

// StateResolveFunction 表示状态解析函数
type StateResolveFunction func(input StateResolveFunctionInput) bool

// ActionFunctionInput 表示操作函数的输入（与StateResolveFunctionInput相同）
type ActionFunctionInput StateResolveFunctionInput

// ActionFunction 表示操作函数，出于性能原因，操作函数会修改输入
type ActionFunction func(input *ActionFunctionInput)
