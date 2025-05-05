package eventreduce

import (
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query_executor"
)

type StateResolverInput struct {
	lq            query.ListeningQuery
	op            db_conn.TransactionOp
	queryExecutor *query_executor.QueryExecutor
}

type StateResolver = func(input StateResolverInput) bool

func ResolveState(input StateResolverInput) StateSet {
	ret := strings.Builder{}
	for _, state := range OrderedStateNames {
		value := GetStateResolverByName(state)(input)
		if value {
			ret.WriteString("1")
		} else {
			ret.WriteString("0")
		}
	}
	return ret.String()
}

func GetStateResolverByName(name StateName) StateResolver {
	switch name {
	case StateHasLimit:
		return HasLimit
	case StateIsFindOne:
		return IsFindOne
	case StateHasSkip:
		return HasSkip
	case StateIsDelete:
		return IsDelete
	case StateIsInsert:
		return IsInsert
	case StateIsUpdate:
		return IsUpdate
	case StateWasResultsEmpty:
		return WasResultEmpty
	case StateWasLimitReached:
		return WasLimitReached
	case StateSortParamsChanged:
		return SortParamsChanged
	case StateWasInResult:
		return WasInResult
	case StateWasFirst:
		return WasFirst
	case StateWasLast:
		return WasLast
	case StateWasSortedBeforeFirst:
		return WasSortedBeforeFirst
	case StateWasSortedAfterLast:
		return WasSortedAfterLast
	case StateIsSortedBeforeFirst:
		return IsSortedBeforeFirst
	case StateIsSortedAfterLast:
		return IsSortedAfterLast
	case StateWasMatching:
		return WasMatching
	case StateDoesMatchNow:
		return DoesMatchNow
	default:
		panic("unknown state")
	}
}

func HasLimit(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if isFindMany {
		return q.Query.Limit > 0
	} else {
		panic("find one query is not supported")
	}
}

func IsFindOne(input StateResolverInput) bool {
	_, isFindOne := input.lq.(*query.FindOneListeningQuery)
	return isFindOne
}

func HasSkip(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if isFindMany {
		return q.Query.Skip > 0
	} else {
		panic("find one query is not supported")
	}
}

func IsDelete(input StateResolverInput) bool {
	_, isDelete := input.op.(*db_conn.DeleteOp)
	return isDelete
}

func IsInsert(input StateResolverInput) bool {
	_, isInsert := input.op.(*db_conn.InsertOp)
	return isInsert
}

func IsUpdate(input StateResolverInput) bool {
	_, isUpdate := input.op.(*db_conn.UpdateOp)
	return isUpdate
}

func WasLimitReached(input StateResolverInput) bool {
	hasLimit := HasLimit(input)
	if !hasLimit {
		return false
	}

	q := input.lq.(*query.FindManyListeningQuery)
	return len(q.Result) >= int(q.Query.Limit)
}

// 如果之前的文档 prevDoc 和当前文档 currDoc 在任意一个排序字段上的
// 值不同，则返回 true
func SortParamsChanged(input StateResolverInput) bool {
	prevDoc := getPrevDoc(input)
	currDoc := getCurrDoc(input)
	if currDoc == nil {
		return false
	}

	if prevDoc == nil {
		return true
	}

	panic("not implemented")
}

func WasInResult(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	opDocId := getDocId(input.op)
	for _, doc := range q.Result {
		if doc.DocId == opDocId {
			return true
		}
	}
	return false
}

func WasFirst(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	if len(q.Result) == 0 {
		return false
	}

	first := q.Result[0]
	return first.DocId == getDocId(input.op)
}

func WasLast(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	if len(q.Result) == 0 {
		return false
	}

	last := q.Result[len(q.Result)-1]
	return last.DocId == getDocId(input.op)
}

func WasSortedBeforeFirst(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	opDocId := getDocId(input.op)
	prevDoc := getPrevDoc(input)

	if len(q.Result) == 0 {
		return false
	}

	first := q.Result[0]
	if first == nil {
		return false
	}

	if first.DocId == opDocId {
		return true
	}

	cmp, err := q.Query.Compare(prevDoc, first.Doc)
	if err != nil {
		log.Warnf("In WasSortedBeforeFirst, compare error: %v", err)
		return false
	}

	return cmp < 0
}

func WasSortedAfterLast(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	opDocId := getDocId(input.op)
	prevDoc := getPrevDoc(input)

	if len(q.Result) == 0 {
		return false
	}

	last := q.Result[len(q.Result)-1]
	if last == nil {
		return false
	}

	if last.DocId == opDocId {
		return true
	}

	cmp, err := q.Query.Compare(prevDoc, last.Doc)
	if err != nil {
		log.Warnf("In WasSortedAfterLast, compare error: %v", err)
		return false
	}

	return cmp > 0
}

func IsSortedBeforeFirst(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	opDocId := getDocId(input.op)
	currDoc := getCurrDoc(input)
	if currDoc == nil {
		return false
	}

	if len(q.Result) == 0 {
		return false
	}

	first := q.Result[0]
	if first.DocId == opDocId {
		return true
	}

	cmp, err := q.Query.Compare(currDoc, first.Doc)
	if err != nil {
		log.Warnf("In IsSortedBeforeFirst, compare error: %v", err)
		return false
	}
	return cmp < 0
}

func IsSortedAfterLast(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	opDocId := getDocId(input.op)
	currDoc := getCurrDoc(input)
	if currDoc == nil {
		return false
	}

	if len(q.Result) == 0 {
		return false
	}

	last := q.Result[len(q.Result)-1]
	if last.DocId == opDocId {
		return true
	}

	cmp, err := q.Query.Compare(currDoc, last.Doc)
	if err != nil {
		log.Warnf("In IsSortedAfterLast, compare error: %v", err)
		return false
	}
	return cmp > 0
}

func WasMatching(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	prevDoc := getPrevDoc(input)
	if prevDoc == nil {
		return false
	}

	match, err := q.Query.Match(prevDoc)
	if err != nil {
		log.Warnf("In WasMatching, match error: %v", err)
		return false
	}
	return match
}

func DoesMatchNow(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	currDoc := getCurrDoc(input)
	if currDoc == nil {
		return false
	}

	match, err := q.Query.Match(currDoc)
	if err != nil {
		log.Warnf("In DoesMatchNow, match error: %v", err)
		return false
	}
	return match
}

func WasResultEmpty(input StateResolverInput) bool {
	q, isFindMany := input.lq.(*query.FindManyListeningQuery)
	if !isFindMany {
		panic("find one query is not supported")
	}

	return len(q.Result) == 0
}

func getCollection(op db_conn.TransactionOp) string {
	switch op := op.(type) {
	case *db_conn.InsertOp:
		return op.Collection
	case *db_conn.UpdateOp:
		return op.Collection
	case *db_conn.DeleteOp:
		return op.Collection
	default:
		panic("unknown transaction op")
	}
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
		panic("unknown transaction op")
	}
}

func getPrevDoc(input StateResolverInput) *loro.LoroDoc {
	opDocId := getDocId(input.op)
	opCollection := getCollection(input.op)
	prevDoc, err := input.queryExecutor.FindOneById(opCollection, opDocId)
	if err != nil {
		log.Warnf("In getPrevDoc, find one by id error: %v", err)
		return nil
	}
	return prevDoc.Doc
}

func getCurrDoc(input StateResolverInput) *loro.LoroDoc {
	switch op := input.op.(type) {
	case *db_conn.InsertOp:
		doc := loro.NewLoroDoc()
		doc.Import(op.Snapshot)
		return doc
	case *db_conn.DeleteOp:
		return nil
	case *db_conn.UpdateOp:
		prevDoc := getPrevDoc(input)
		if prevDoc == nil {
			return nil
		}
		forked := prevDoc.Fork()
		forked.Import(op.Update)
		return forked
	default:
		panic("unknown transaction op")
	}
}
