package synchronizer2

import (
	"sync"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/eventreduce"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permission_proxy"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query_executor"
)

// QueryManager is responsible for managing all the queries subscribed by clients
//
// It also responsible for keeping the result of the query up to date, by
// listening to the incoming transactions, using the EventReduce algorithm to
// calculate the Action to take for updating the result set, and then executing
// the ActionFunction to update the result set.
type QueryManager struct {
	// Queries subscribed by each client
	// clientId -> queryHash -> query
	subscriptions   map[string]map[string]query.ListeningQuery
	queryExecutor   *query_executor.QueryExecutor
	permissionProxy *permission_proxy.PermissionProxy
	eventReducer    eventreduce.EventReducer
	mu              sync.RWMutex // Protect concurrent access to subscriptions
}

// NewQueryManager creates and returns a new QueryManager instance
func NewQueryManager(queryExecutor *query_executor.QueryExecutor, permissionProxy *permission_proxy.PermissionProxy) *QueryManager {
	return &QueryManager{
		subscriptions:   make(map[string]map[string]query.ListeningQuery),
		queryExecutor:   queryExecutor,
		permissionProxy: permissionProxy,
		eventReducer:    eventreduce.GetEventReducer(),
		mu:              sync.RWMutex{},
	}
}

// createListeningQuery creates a ListeningQuery instance, executes the query, and stores the result in Result
func (m *QueryManager) createListeningQuery(q query.Query) (query.ListeningQuery, error) {
	switch q := q.(type) {
	case *query.FindOneQuery:
		res, err := m.queryExecutor.FindOne(q)
		if err != nil {
			return nil, err
		}
		return &query.FindOneListeningQuery{
			Query:  q,
			Error:  nil,
			Result: res,
		}, nil
	case *query.FindManyQuery:
		res, err := m.queryExecutor.FindMany(q)
		if err != nil {
			return nil, err
		}
		return &query.FindManyListeningQuery{
			Query:  q,
			Error:  nil,
			Result: res,
		}, nil
	default:
		panic("unknown query type")
	}
}

// SubscribeNewQuery subscribes to a new query
func (s *QueryManager) SubscribeNewQuery(clientId string, newQuery query.Query) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	ss, ok := s.subscriptions[clientId]
	if !ok {
		ss = make(map[string]query.ListeningQuery)
		s.subscriptions[clientId] = ss
	}

	queryHash, err := query.StableStringify(newQuery)
	if err != nil {
		return err
	}

	lq, err := s.createListeningQuery(newQuery)
	if err != nil {
		return err
	}
	ss[queryHash] = lq

	return nil
}

// RemoveSubscriptedQuery removes the query subscription for the specified client
func (s *QueryManager) RemoveSubscriptedQuery(clientId string, q query.Query) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	queryHash, err := query.StableStringify(q)
	if err != nil {
		return err
	}

	if clientMap, ok := s.subscriptions[clientId]; ok {
		delete(clientMap, queryHash)
	}
	return nil
}

// CheckSubscriptedQuery checks if the specified client has subscribed to the given query
func (a *QueryManager) CheckSubscriptedQuery(clientId string, q query.Query) (bool, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	queryHash, err := query.StableStringify(q)
	if err != nil {
		return false, err
	}

	clientMap, ok := a.subscriptions[clientId]
	if !ok {
		return false, nil
	}
	return clientMap[queryHash] != nil, nil
}

type ClientUpdates struct {
	Updates map[string][]byte
	Deletes map[string]struct{}
}

func (cu *ClientUpdates) IsEmpty() bool {
	return len(cu.Updates) == 0 && len(cu.Deletes) == 0
}

func (a *QueryManager) HandleTransaction(txn *db_conn.Transaction) map[string]*ClientUpdates {
	a.mu.RLock()
	defer a.mu.RUnlock()

	cu := make(map[string]*ClientUpdates)
	for _, op := range txn.Operations {
		for clientId, queries := range a.subscriptions {
			clientUpdates, ok := cu[clientId]
			if !ok {
				clientUpdates = &ClientUpdates{
					Updates: make(map[string][]byte),
					Deletes: make(map[string]struct{}),
				}
				cu[clientId] = clientUpdates
			}

			for _, lq := range queries {
				// Use the EventReduce algorithm to calculate the Action to take for updating the result set
				action := a.eventReducer.Reduce(lq, op)
				// Get the corresponding ActionFunction based on the Action
				actionFunc := GetActionFunction(action)
				// Execute the ActionFunction
				actionFunc(ActionFunctionInput{
					clientId:       clientId,
					permissions:    a.permissionProxy,
					listeningQuery: lq,
					op:             op,
					clientUpdates:  clientUpdates,
					queryExecutor:  a.queryExecutor,
				})
			}
		}
	}
	return cu
}
