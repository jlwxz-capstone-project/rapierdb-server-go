package client

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
)

const (
	// QUERY_STATUS_LOADING 表示查询正在加载中
	QUERY_STATUS_LOADING QueryStatus = 0
	// QUERY_STATUS_LOCAL 表示查询结果来自本地数据库，未与服务器同步
	QUERY_STATUS_LOCAL QueryStatus = 1
	// QUERY_STATUS_SYNCED 表示查询结果已与服务器同步
	QUERY_STATUS_SYNCED QueryStatus = 2
	// QUERY_STATUS_ERROR 表示查询发生错误
	QUERY_STATUS_ERROR QueryStatus = 3
)

type QueryStatus int

type ReactiveQuery interface {
	isReactiveQuery()
}

type DocWithId struct {
	DocId string
	Doc   *loro.LoroDoc
}

type ReactiveFindOneQuery struct {
	Query  *query.FindOneQuery
	Status *Ref[QueryStatus]
	// 如果 Status 为 QUERY_STATUS_ERROR，则 Error 为错误信息
	Error *Ref[error]
	// Result 如果没有匹配的文档，或者查询未成功，则为 nil
	Result *Ref[*DocWithId]
}

func (q *ReactiveFindOneQuery) isReactiveQuery() {}

type ReactiveFindManyQuery struct {
	Query  *query.FindManyQuery
	Status *Ref[QueryStatus]
	// 如果 Status 为 QUERY_STATUS_ERROR，则 Error 为错误信息
	Error *Ref[error]
	// Result 总是一个数组，不可能为 nil
	Result *Ref[[]*DocWithId]
}

func (q *ReactiveFindManyQuery) isReactiveQuery() {}
