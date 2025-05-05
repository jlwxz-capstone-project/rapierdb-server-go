package synchronizer2

import (
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/eventreduce"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/key_utils"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permission_proxy"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query_executor"
)

type ActionFunctionInput struct {
	clientId       string
	listeningQuery query.ListeningQuery
	permissions    *permission_proxy.PermissionProxy
	op             db_conn.TransactionOp
	clientUpdates  *ClientUpdates
	queryExecutor  *query_executor.QueryExecutor
}

type ActionFunction func(in ActionFunctionInput)

func GetActionFunction(actionName eventreduce.ActionName) ActionFunction {
	switch actionName {
	case eventreduce.ActionDoNothing:
		return ActionDoNothing
	case eventreduce.ActionInsertFirst:
		return ActionInsertFirst
	case eventreduce.ActionInsertLast:
		return ActionInsertLast
	case eventreduce.ActionRemoveFirstItem:
		return ActionRemoveFirstItem
	case eventreduce.ActionRemoveLastItem:
		return ActionRemoveLastItem
	case eventreduce.ActionRemoveFirstInsertLast:
		return ActionRemoveFirstInsertLast
	case eventreduce.ActionRemoveLastInsertFirst:
		return ActionRemoveLastInsertFirst
	case eventreduce.ActionRemoveFirstInsertFirst:
		return ActionRemoveFirstInsertFirst
	case eventreduce.ActionRemoveLastInsertLast:
		return ActionRemoveLastInsertLast
	case eventreduce.ActionRemoveExisting:
		return ActionRemoveExisting
	case eventreduce.ActionReplaceExisting:
		return ActionReplaceExisting
	case eventreduce.ActionInsertAtSortPosition:
		return ActionInsertAtSortPosition
	case eventreduce.ActionRemoveExistingAndInsertAtSortPosition:
		return ActionRemoveExistingAndInsertAtSortPosition
	case eventreduce.ActionRunFullQueryAgain:
		return ActionRunFullQueryAgain
	default:
		panic(fmt.Sprintf("unsupport action: %v", actionName))
	}
}

func updateClientUpdates(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	switch op := in.op.(type) {
	case *db_conn.InsertOp:
		key, err := key_utils.CalcDocKey(lq.Query.Collection, op.DocID)
		if err != nil {
			panic(fmt.Sprintf("calc doc key error: %v", err))
		}
		stringKey := string(key)
		// 如果该键已在删除集合中，先从删除集合中移除（虽然一般不会出现这种情况）
		delete(in.clientUpdates.Deletes, stringKey)
		in.clientUpdates.Updates[stringKey] = op.Snapshot
	case *db_conn.UpdateOp:
		key, err := key_utils.CalcDocKey(lq.Query.Collection, op.DocID)
		if err != nil {
			panic(fmt.Sprintf("calc doc key error: %v", err))
		}
		stringKey := string(key)
		// 如果该键已在删除集合中，则不应再更新它
		if _, exists := in.clientUpdates.Deletes[stringKey]; !exists {
			in.clientUpdates.Updates[stringKey] = op.Update
		}
	case *db_conn.DeleteOp:
		key, err := key_utils.CalcDocKey(lq.Query.Collection, op.DocID)
		if err != nil {
			panic(fmt.Sprintf("calc doc key error: %v", err))
		}
		stringKey := string(key)
		// 删除操作具有最高优先级，从更新映射中移除该键并添加到删除集合中
		delete(in.clientUpdates.Updates, stringKey)
		in.clientUpdates.Deletes[stringKey] = struct{}{}
	}
}

func ActionDoNothing(in ActionFunctionInput) {
}

func ActionInsertFirst(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	insertOp, isInsertOp := in.op.(*db_conn.InsertOp)
	if isFindMany {
		if isInsertOp {
			doc := loro.NewLoroDoc()
			doc.Import(insertOp.Snapshot)
			docWithId := &query.DocWithId{
				DocId: insertOp.DocID,
				Doc:   doc,
			}
			lq.Result = append([]*query.DocWithId{docWithId}, lq.Result...)
			updateClientUpdates(in)
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionInsertLast(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	insertOp, isInsertOp := in.op.(*db_conn.InsertOp)
	if isFindMany {
		if isInsertOp {
			doc := loro.NewLoroDoc()
			doc.Import(insertOp.Snapshot)
			docWithId := &query.DocWithId{
				DocId: insertOp.DocID,
				Doc:   doc,
			}
			lq.Result = append(lq.Result, docWithId)
			updateClientUpdates(in)
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveFirstItem(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	if isFindMany {
		if len(lq.Result) > 0 {
			lq.Result = lq.Result[1:]
			updateClientUpdates(in)
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveLastItem(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	if isFindMany {
		if len(lq.Result) > 0 {
			lq.Result = lq.Result[:len(lq.Result)-1]
			updateClientUpdates(in)
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveFirstInsertLast(in ActionFunctionInput) {
	ActionRemoveFirstItem(in)
	ActionInsertLast(in)
}

func ActionRemoveLastInsertFirst(in ActionFunctionInput) {
	ActionRemoveLastItem(in)
	ActionInsertFirst(in)
}

func ActionRemoveFirstInsertFirst(in ActionFunctionInput) {
	ActionRemoveFirstItem(in)
	ActionInsertFirst(in)
}

func ActionRemoveLastInsertLast(in ActionFunctionInput) {
	ActionRemoveLastItem(in)
	ActionInsertLast(in)
}

func ActionRemoveExisting(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	// XXX 二分搜索
	docId := getDocId(in.op)
	idx := -1
	for i, doc := range lq.Result {
		if doc.DocId == docId {
			idx = i
			break
		}
	}

	if idx != -1 {
		lq.Result = append(lq.Result[:idx], lq.Result[idx+1:]...)
		updateClientUpdates(in)
	}
}

func ActionReplaceExisting(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	op, isUpdateOp := in.op.(*db_conn.UpdateOp)
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
				updateClientUpdates(in)
			}
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionAlwaysWrong(in ActionFunctionInput) {
	panic("not implemented")
}

func ActionInsertAtSortPosition(in ActionFunctionInput) {
	lq, isFindMany := in.listeningQuery.(*query.FindManyListeningQuery)
	insertOp, isInsertOp := in.op.(*db_conn.InsertOp)
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
			updateClientUpdates(in)
		} else {
			panic("unexpected operation")
		}
	} else {
		panic("find one query is not supported")
	}
}

func ActionRemoveExistingAndInsertAtSortPosition(in ActionFunctionInput) {
	ActionRemoveExisting(in)
	ActionInsertAtSortPosition(in)
}

func ActionRunFullQueryAgain(in ActionFunctionInput) {
	switch lq := in.listeningQuery.(type) {
	case *query.FindOneListeningQuery:
		panic("find one query is not supported")
	case *query.FindManyListeningQuery:
		res, err := in.queryExecutor.FindMany(lq.Query)
		if err != nil {
			panic(fmt.Sprintf("find many error: %v", err))
		}
		lq.Result = res
		updateClientUpdates(in)
	}
}

func ActionUnknownAction(in ActionFunctionInput) {
	panic("unknown action")
}

func getDocId(op db_conn.TransactionOp) string {
	switch op := op.(type) {
	case *db_conn.InsertOp:
		return op.DocID
	case *db_conn.UpdateOp:
		return op.DocID
	case *db_conn.DeleteOp:
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
