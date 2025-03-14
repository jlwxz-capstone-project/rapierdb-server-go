package client

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
)

type ReactiveQuery interface {
	isReactiveQuery()
}

type ReactiveFindOneQuery struct {
	Query     *query.FindOneQuery
	IsLoading *Ref[bool]
	Error     *Ref[error]
	Result    *Ref[struct {
		DocId string
		Doc   *loro.LoroDoc
	}]
}

func (q *ReactiveFindOneQuery) isReactiveQuery() {}

type ReactiveFindManyQuery struct {
	Query     *query.FindManyQuery
	IsLoading *Ref[bool]
	Error     *Ref[error]
	Result    *Ref[[]struct {
		DocId string
		Doc   *loro.LoroDoc
	}]
}

func (q *ReactiveFindManyQuery) isReactiveQuery() {}
