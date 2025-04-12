package query

import (
	"fmt"
	"sort"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
)

type DocWithId struct {
	DocId string
	Doc   *loro.LoroDoc
}

type FindOneResult = *DocWithId
type FindManyResult = []*DocWithId

type QueryExecutor struct {
	StorageEngine *storage_engine.StorageEngine
}

func NewQueryExecutor(storageEngine *storage_engine.StorageEngine) *QueryExecutor {
	return &QueryExecutor{
		StorageEngine: storageEngine,
	}
}

func (q *QueryExecutor) FindOne(query *FindOneQuery) (FindOneResult, error) {
	// TODO: 全部载入内存肯定是不好的
	docs, err := q.StorageEngine.LoadAllDocsInCollection(query.Collection, true)
	if err != nil {
		return nil, err
	}

	for docId, doc := range docs {
		ok, err := query.Match(doc)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		if ok {
			return &DocWithId{
				DocId: docId,
				Doc:   doc,
			}, nil
		}
	}
	return nil, nil
}

func (q *QueryExecutor) FindMany(query *FindManyQuery) (FindManyResult, error) {
	// TODO: 全部载入内存肯定是不好的
	docs, err := q.StorageEngine.LoadAllDocsInCollection(query.Collection, true)
	if err != nil {
		return nil, err
	}

	// 过滤出与查询条件匹配的文档
	result := make(FindManyResult, 0)
	for docId, doc := range docs {
		ok, err := query.Match(doc)
		if err != nil {
			fmt.Printf("%+v\n", err)
		}
		if ok {
			result = append(result, &DocWithId{
				DocId: docId,
				Doc:   doc,
			})
		}
	}

	// 处理排序
	if len(query.Sort) > 0 {
		sort.Slice(result, func(i, j int) bool {
			cmp, err := query.Compare(result[i].Doc, result[j].Doc)
			if err != nil {
				// 排序出错时，保持原顺序
				return i < j
			}
			return cmp < 0 // 小于0表示i应该排在j前面
		})
	}

	// 处理 Skip
	if query.Skip > 0 {
		if int64(len(result)) <= query.Skip {
			// 如果跳过的数量大于等于结果集大小，返回空结果
			return FindManyResult{}, nil
		}
		result = result[query.Skip:]
	}

	// 处理 Limit
	if query.Limit > 0 && int64(len(result)) > query.Limit {
		result = result[:query.Limit]
	}

	return result, nil
}
