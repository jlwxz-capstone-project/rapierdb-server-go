package query_executor

import (
	"fmt"
	"sort"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query/doc_visitor"
)

type QueryExecutor struct {
	conn db_conn.DbConnection
}

func NewQueryExecutor(dbConn db_conn.DbConnection) *QueryExecutor {
	return &QueryExecutor{
		conn: dbConn,
	}
}

func (q *QueryExecutor) FindOneById(collection string, id string) (query.FindOneResult, error) {
	// TODO: 全部载入内存肯定是不好的
	docs, err := q.conn.LoadCollection(collection)
	if err != nil {
		return nil, err
	}

	doc, ok := docs[id]
	if !ok {
		return nil, nil
	}

	if doc_visitor.IsDeleted(doc) {
		return nil, nil
	}

	return &query.DocWithId{
		DocId: id,
		Doc:   doc,
	}, nil
}

func (qe *QueryExecutor) FindOne(q *query.FindOneQuery) (query.FindOneResult, error) {
	// TODO: load all docs into memory is not good
	docs, err := qe.conn.LoadCollection(q.Collection)
	if err != nil {
		return nil, err
	}

	for docId, doc := range docs {
		ok, err := q.Match(doc)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		if ok {
			return &query.DocWithId{
				DocId: docId,
				Doc:   doc,
			}, nil
		}
	}
	return nil, nil
}

func (qe *QueryExecutor) FindMany(q *query.FindManyQuery) (query.FindManyResult, error) {
	// TODO: load all docs into memory is not good
	docs, err := qe.conn.LoadCollection(q.Collection)
	if err != nil {
		return nil, err
	}

	// filter out docs that match the query
	result := make(query.FindManyResult, 0)
	for docId, doc := range docs {
		ok, err := q.Match(doc)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		if ok {
			result = append(result, &query.DocWithId{
				DocId: docId,
				Doc:   doc,
			})
		}
	}

	// handle sorting
	if len(q.Sort) > 0 {
		sort.Slice(result, func(i, j int) bool {
			cmp, err := q.Compare(result[i].Doc, result[j].Doc)
			if err != nil {
				// when sorting error occurs, keep the original order
				return i < j
			}
			return cmp < 0 // cmp < 0 means i should be before j
		})
	} else {
		// if no sorting is specified, sort by doc id (primary key)
		// this is very important, because EventReduce algorithm depends on the order of documents in the result
		sort.Slice(result, func(i, j int) bool {
			return result[i].DocId < result[j].DocId
		})
	}

	// handle skip
	if q.Skip > 0 {
		if int64(len(result)) <= q.Skip {
			// if the number of skipped docs is greater than or equal to the result size, return empty result
			return query.FindManyResult{}, nil
		}
		result = result[q.Skip:]
	}

	// handle limit
	if q.Limit > 0 && int64(len(result)) > q.Limit {
		result = result[:q.Limit]
	}

	return result, nil
}

func (qe *QueryExecutor) IsValidCollection(collection string) bool {
	dbMeta := qe.conn.GetDatabaseMeta()
	for _, c := range dbMeta.GetCollectionNames() {
		if c == collection {
			return true
		}
	}
	return false
}
