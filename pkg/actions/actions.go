package actions

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// 按性能开销排序的操作列表，最低开销在前
var OrderedActionList = []types.ActionName{
	types.ActionDoNothing,
	types.ActionInsertFirst,
	types.ActionInsertLast,
	types.ActionRemoveFirstItem,
	types.ActionRemoveLastItem,
	types.ActionRemoveFirstInsertLast,
	types.ActionRemoveLastInsertFirst,
	types.ActionRemoveFirstInsertFirst,
	types.ActionRemoveLastInsertLast,
	types.ActionRemoveExisting,
	types.ActionReplaceExisting,
	types.ActionAlwaysWrong,
	types.ActionInsertAtSortPosition,
	types.ActionRemoveExistingAndInsertAtSortPosition,
	types.ActionRunFullQueryAgain,
	types.ActionUnknownAction,
}

// 操作函数映射
var ActionFunctions = map[types.ActionName]types.ActionFunction{
	types.ActionDoNothing:                             DoNothing,
	types.ActionInsertFirst:                           InsertFirst,
	types.ActionInsertLast:                            InsertLast,
	types.ActionRemoveFirstItem:                       RemoveFirstItem,
	types.ActionRemoveLastItem:                        RemoveLastItem,
	types.ActionRemoveFirstInsertLast:                 RemoveFirstInsertLast,
	types.ActionRemoveLastInsertFirst:                 RemoveLastInsertFirst,
	types.ActionRemoveFirstInsertFirst:                RemoveFirstInsertFirst,
	types.ActionRemoveLastInsertLast:                  RemoveLastInsertLast,
	types.ActionRemoveExisting:                        RemoveExisting,
	types.ActionReplaceExisting:                       ReplaceExisting,
	types.ActionAlwaysWrong:                           AlwaysWrong,
	types.ActionInsertAtSortPosition:                  InsertAtSortPosition,
	types.ActionRemoveExistingAndInsertAtSortPosition: RemoveExistingAndInsertAtSortPosition,
	types.ActionRunFullQueryAgain:                     RunFullQueryAgain,
	types.ActionUnknownAction:                         UnknownAction,
}
