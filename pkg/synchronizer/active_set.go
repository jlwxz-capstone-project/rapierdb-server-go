package synchronizer

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/types"
)

// 一个 EventReducer 会根据 ListeningQuery lq 和 TransactionOp
// 计算出应该采取什么 Action 更新 lq 的结果集
type EventReducer interface {
	Reduce(lq ListeningQuery, op storage_engine.TransactionOp) types.ActionName
}

type QueryManager struct {
	// 每个客户端订阅的查询
	// clientId -> queryHash -> query
	subscriptions map[string]map[string]ListeningQuery
	queryExecutor *query.QueryExecutor
	eventReducer  EventReducer
}

// NewQueryManager 创建并返回一个新的 QueryManager 实例
func NewQueryManager(queryExecutor *query.QueryExecutor) *QueryManager {
	return &QueryManager{
		subscriptions: make(map[string]map[string]ListeningQuery),
		queryExecutor: queryExecutor,
	}
}

// createListeningQuery 创建一个 ListeningQuery 实例，会执行查询放到 Result 中
func (m *QueryManager) createListeningQuery(q query.Query) (ListeningQuery, error) {
	switch q := q.(type) {
	case *query.FindOneQuery:
		res, err := m.queryExecutor.FindOne(q)
		if err != nil {
			return nil, err
		}
		return &FindOneListeningQuery{
			Query:  q,
			Error:  nil,
			Result: res,
		}, nil
	case *query.FindManyQuery:
		res, err := m.queryExecutor.FindMany(q)
		if err != nil {
			return nil, err
		}
		return &FindManyListeningQuery{
			Query:  q,
			Error:  nil,
			Result: res,
		}, nil
	default:
		panic("unknown query type")
	}
}

// SubscribeNewQuery 订阅新的查询
func (s *QueryManager) SubscribeNewQuery(clientId string, newQuery query.Query) error {
	ss, ok := s.subscriptions[clientId]
	if !ok {
		ss = make(map[string]ListeningQuery)
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

// RemoveSubscriptedQuery 移除指定客户端的查询订阅
func (s *QueryManager) RemoveSubscriptedQuery(clientId string, q query.Query) error {
	queryHash, err := query.StableStringify(q)
	if err != nil {
		return err
	}

	if clientMap, ok := s.subscriptions[clientId]; ok {
		delete(clientMap, queryHash)
	}
	return nil
}

// CheckSubscriptedQuery 检查指定客户端是否订阅了给定的查询
func (a *QueryManager) CheckSubscriptedQuery(clientId string, q query.Query) (bool, error) {
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

func (a *QueryManager) HandleTransaction(txn *storage_engine.Transaction) map[string]*ClientUpdates {
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
				// 使用 EventReduce 算法计算更新结果集应该采取的 Action
				action := a.eventReducer.Reduce(lq, op)
				// 根据 Action 获取对应的 ActionFunction
				actionFunc := GetActionFunction(action)
				// 执行 ActionFunction
				actionFunc(ActionFunctionInput{
					listeningQuery: lq,
					queryExecutor:  a.queryExecutor,
					op:             op,
					clientUpdates:  clientUpdates,
				})
			}
		}
	}
	return cu
}
