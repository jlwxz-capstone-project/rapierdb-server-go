package synchronizer

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
)

type ListeningQuery interface {
	isListeningQuery()
	GetQuery() query.Query
}

type FindOneListeningQuery struct {
	Query  *query.FindOneQuery
	Error  *error
	Result query.FindOneResult
}

func (r *FindOneListeningQuery) isListeningQuery()     {}
func (r *FindOneListeningQuery) GetQuery() query.Query { return r.Query }

type FindManyListeningQuery struct {
	Query  *query.FindManyQuery
	Error  *error
	Result query.FindManyResult
}

func (r *FindManyListeningQuery) isListeningQuery()     {}
func (r *FindManyListeningQuery) GetQuery() query.Query { return r.Query }
