package query

import (
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

func (q *QueryExecutor) Execute(collection string, query *Query) (map[string]*loro.LoroDoc, error) {
	// 全部载入内存有问题
	docs, err := q.StorageEngine.LoadAllDocsInCollection(collection, true)
	if err != nil {
		return nil, err
	}

	result := make(map[string]*loro.LoroDoc)
	for docId, doc := range docs {
		// 顺序扫描性能不好
		ok, err := query.Match(doc)
		if err != nil {
			return nil, err
		}
		if ok {
			result[docId] = doc
		}
	}
	return result, nil
}
