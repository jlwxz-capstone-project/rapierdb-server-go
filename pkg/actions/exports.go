package actions

import "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"

// 导出所有操作函数
var (
	DoNothingFunc                             types.ActionFunction = DoNothing
	InsertFirstFunc                           types.ActionFunction = InsertFirst
	InsertLastFunc                            types.ActionFunction = InsertLast
	RemoveFirstItemFunc                       types.ActionFunction = RemoveFirstItem
	RemoveLastItemFunc                        types.ActionFunction = RemoveLastItem
	RemoveFirstInsertLastFunc                 types.ActionFunction = RemoveFirstInsertLast
	RemoveLastInsertFirstFunc                 types.ActionFunction = RemoveLastInsertFirst
	RemoveFirstInsertFirstFunc                types.ActionFunction = RemoveFirstInsertFirst
	RemoveLastInsertLastFunc                  types.ActionFunction = RemoveLastInsertLast
	RemoveExistingFunc                        types.ActionFunction = RemoveExisting
	ReplaceExistingFunc                       types.ActionFunction = ReplaceExisting
	AlwaysWrongFunc                           types.ActionFunction = AlwaysWrong
	InsertAtSortPositionFunc                  types.ActionFunction = InsertAtSortPosition
	RemoveExistingAndInsertAtSortPositionFunc types.ActionFunction = RemoveExistingAndInsertAtSortPosition
	RunFullQueryAgainFunc                     types.ActionFunction = RunFullQueryAgain
	UnknownActionFunc                         types.ActionFunction = UnknownAction

	// 工具函数
	PushAtSortPositionFunc = PushAtSortPosition
)

// 导出操作列表和函数映射
var (
	ActionList    = OrderedActionList
	ActionFuncMap = ActionFunctions
)
