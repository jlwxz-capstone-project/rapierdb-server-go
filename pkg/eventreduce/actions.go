package eventreduce

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

var OrderedActions = []ActionName{
	ActionDoNothing,
	ActionInsertFirst,
	ActionInsertLast,
	ActionRemoveFirstItem,
	ActionRemoveLastItem,
	ActionRemoveFirstInsertLast,
	ActionRemoveLastInsertFirst,
	ActionRemoveFirstInsertFirst,
	ActionRemoveLastInsertLast,
	ActionRemoveExisting,
	ActionReplaceExisting,
	ActionAlwaysWrong,
	ActionInsertAtSortPosition,
	ActionRemoveExistingAndInsertAtSortPosition,
	ActionRunFullQueryAgain,
	ActionUnknownAction,
}
