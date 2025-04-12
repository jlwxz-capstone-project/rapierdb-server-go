package synchronizer

import (
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

type ActionFunctionInput struct {
	clientId       string
	listeningQuery ListeningQuery
	permissions    *query.Permissions
	op             storage_engine.TransactionOp
	clientUpdates  *ClientUpdates
}

type ActionFunction func(input ActionFunctionInput)

func GetActionFunction(actionName types.ActionName) ActionFunction {
	switch actionName {
	case types.ActionDoNothing:
		return ActionDoNothing
	case types.ActionInsertFirst:
		return ActionInsertFirst
	case types.ActionInsertLast:
		return ActionInsertLast
	case types.ActionRemoveFirstItem:
		return ActionRemoveFirstItem
	case types.ActionRemoveLastItem:
		return ActionRemoveLastItem
	case types.ActionRemoveFirstInsertLast:
		return ActionRemoveFirstInsertLast
	case types.ActionRemoveLastInsertFirst:
		return ActionRemoveLastInsertFirst
	case types.ActionRemoveFirstInsertFirst:
		return ActionRemoveFirstInsertFirst
	case types.ActionRemoveLastInsertLast:
		return ActionRemoveLastInsertLast
	case types.ActionRemoveExisting:
		return ActionRemoveExisting
	case types.ActionReplaceExisting:
		return ActionReplaceExisting
	case types.ActionInsertAtSortPosition:
		return ActionInsertAtSortPosition
	case types.ActionRemoveExistingAndInsertAtSortPosition:
		return ActionRemoveExistingAndInsertAtSortPosition
	case types.ActionRunFullQueryAgain:
		return ActionRunFullQueryAgain
	default:
		panic(fmt.Sprintf("unsupport action: %v", actionName))
	}
}

func ActionDoNothing(input ActionFunctionInput) {
}

func ActionInsertFirst(input ActionFunctionInput) {
	lq, isFindMany := input.listeningQuery.(*FindManyListeningQuery)
	insertOp, isInsertOp := input.op.(*storage_engine.InsertOp)
	if isFindMany {
		if isInsertOp {
			doc := loro.NewLoroDoc()
			doc.Import(insertOp.Snapshot)
			docWithId := &query.DocWithId{
				DocId: insertOp.DocID,
				Doc:   doc,
			}
			lq.Result = append([]*query.DocWithId{docWithId}, lq.Result...)
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionInsertLast(input ActionFunctionInput) {
	lq, isFindMany := input.listeningQuery.(*FindManyListeningQuery)
	insertOp, isInsertOp := input.op.(*storage_engine.InsertOp)
	if isFindMany {
		if isInsertOp {
			doc := loro.NewLoroDoc()
			doc.Import(insertOp.Snapshot)
			docWithId := &query.DocWithId{
				DocId: insertOp.DocID,
				Doc:   doc,
			}
			lq.Result = append(lq.Result, docWithId)
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveFirstItem(input ActionFunctionInput) {
	lq, isFindMany := input.listeningQuery.(*FindManyListeningQuery)
	if isFindMany {
		if len(lq.Result) > 0 {
			lq.Result = lq.Result[1:]
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveLastItem(input ActionFunctionInput) {
	lq, isFindMany := input.listeningQuery.(*FindManyListeningQuery)
	if isFindMany {
		if len(lq.Result) > 0 {
			lq.Result = lq.Result[:len(lq.Result)-1]
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveFirstInsertLast(input ActionFunctionInput) {
	ActionRemoveFirstItem(input)
	ActionInsertLast(input)
}

func ActionRemoveLastInsertFirst(input ActionFunctionInput) {
	ActionRemoveLastItem(input)
	ActionInsertFirst(input)
}

func ActionRemoveFirstInsertFirst(input ActionFunctionInput) {
	ActionRemoveFirstItem(input)
	ActionInsertFirst(input)
}

func ActionRemoveLastInsertLast(input ActionFunctionInput) {
	ActionRemoveLastItem(input)
	ActionInsertLast(input)
}

func ActionRemoveExisting(input ActionFunctionInput) {
	lq, isFindMany := input.listeningQuery.(*FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	// XXX 二分搜索
	docId := getDocId(input.op)
	idx := -1
	for i, doc := range lq.Result {
		if doc.DocId == docId {
			idx = i
			break
		}
	}

	if idx != -1 {
		lq.Result = append(lq.Result[:idx], lq.Result[idx+1:]...)
	}
}

func ActionReplaceExisting(input ActionFunctionInput) {
	lq, isFindMany := input.listeningQuery.(*FindManyListeningQuery)
	op, isUpdateOp := input.op.(*storage_engine.UpdateOp)
	if isFindMany {
		if isUpdateOp {
			idx := -1
			for i, doc := range lq.Result {
				if doc.DocId == op.DocID {
					idx = i
					break
				}
			}

			if idx != -1 {
				lq.Result[idx].Doc.Import(op.Update)
			}
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionAlwaysWrong(input ActionFunctionInput) {
	panic("not implemented")
}

func ActionInsertAtSortPosition(input ActionFunctionInput) {
	lq, isFindMany := input.listeningQuery.(*FindManyListeningQuery)
	insertOp, isInsertOp := input.op.(*storage_engine.InsertOp)
	if isFindMany {
		if isInsertOp {
			doc := loro.NewLoroDoc()
			doc.Import(insertOp.Snapshot)
			docWithId := &query.DocWithId{
				DocId: insertOp.DocID,
				Doc:   doc,
			}
			cmp := func(doc1, doc2 *query.DocWithId) int {
				cmp, err := lq.Query.Compare(doc1.Doc, doc2.Doc)
				if err != nil {
					panic(fmt.Sprintf("compare error: %v", err))
				}
				return cmp
			}
			pushAtSortPos(&lq.Result, docWithId, cmp, 0)
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveExistingAndInsertAtSortPosition(input ActionFunctionInput) {
	ActionRemoveExisting(input)
	ActionInsertAtSortPosition(input)
}

func ActionRunFullQueryAgain(input ActionFunctionInput) {
	panic("not implemented")
}

func ActionUnknownAction(input ActionFunctionInput) {
	panic("not implemented")
}

func getDocId(op storage_engine.TransactionOp) string {
	switch op := op.(type) {
	case *storage_engine.InsertOp:
		return op.DocID
	case *storage_engine.UpdateOp:
		return op.DocID
	case *storage_engine.DeleteOp:
		return op.DocID
	default:
		panic("unexpected operation")
	}
}

// ref: https://www.npmjs.com/package/array-push-at-sort-position
func pushAtSortPos[T any](arr *[]T, item T, cmp func(T, T) int, low int) int {
	length := len(*arr)

	high := length - 1
	mid := 0

	if length == 0 {
		*arr = append(*arr, item)
		return 0
	}

	var lastMidItem T

	for low <= high {
		mid := int(low + (high-low)/2)
		lastMidItem = (*arr)[mid]
		if cmp(lastMidItem, item) <= 0 {
			low = mid + 1
		} else {
			high = mid - 1
		}
	}

	if cmp(lastMidItem, item) <= 0 {
		mid++
	}

	newArr := make([]T, length+1)
	copy(newArr[:mid], (*arr)[:mid])
	newArr[mid] = item
	copy(newArr[mid+1:], (*arr)[mid:])
	*arr = append([]T{}, newArr...)

	return mid
}
