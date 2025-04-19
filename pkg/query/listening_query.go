package query

type ListeningQuery interface {
	isListeningQuery()
	GetQuery() Query
}

type FindOneListeningQuery struct {
	Query  *FindOneQuery
	Error  *error
	Result FindOneResult
}

func (r *FindOneListeningQuery) isListeningQuery() {}
func (r *FindOneListeningQuery) GetQuery() Query   { return r.Query }

type FindManyListeningQuery struct {
	Query  *FindManyQuery
	Error  *error
	Result FindManyResult
}

func (r *FindManyListeningQuery) isListeningQuery() {}
func (r *FindManyListeningQuery) GetQuery() Query   { return r.Query }
