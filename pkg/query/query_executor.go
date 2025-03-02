package query

import (
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
)

type QueryExecutor struct {
	StorageEngine *storage_engine.StorageEngine
}

func NewQueryExecutor(storageEngine *storage_engine.StorageEngine) *QueryExecutor {
	return &QueryExecutor{
		StorageEngine: storageEngine,
	}
}

func (q *QueryExecutor) FindOne(collection string, query *FindOneQuery) (*loro.LoroDoc, error) {
	// 全部载入内存有问题
	docs, err := q.StorageEngine.LoadAllDocsInCollection(collection, true)
	if err != nil {
		return nil, err
	}

	for _, doc := range docs {
		ok, err := query.Match(doc)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		if ok {
			return doc, nil
		}
	}
	return nil, nil
}

func (q *QueryExecutor) FindMany(collection string, query *FindManyQuery) ([]*loro.LoroDoc, error) {
	docs, err := q.StorageEngine.LoadAllDocsInCollection(collection, true)
	if err != nil {
		return nil, err
	}

	result := make([]*loro.LoroDoc, 0)
	for _, doc := range docs {
		ok, err := query.Match(doc)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		if ok {
			result = append(result, doc)
		}
	}
	return result, nil
}
