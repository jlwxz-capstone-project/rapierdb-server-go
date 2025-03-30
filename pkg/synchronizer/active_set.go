package synchronizer

import "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"

// ActiveSet 是一个活跃集合，用于
// 存储每个客户端订阅的查询
type ActiveSet struct {
	// clientId -> queryHash -> query
	subscriptions map[string]map[string]query.Query
}

// NewActiveSet 创建并返回一个新的 ActiveSet 实例
func NewActiveSet() *ActiveSet {
	return &ActiveSet{
		subscriptions: make(map[string]map[string]query.Query),
	}
}

// UpdateSubscription 更新指定客户端的查询订阅
func (s *ActiveSet) UpdateSubscription(clientId string, newQueries []query.Query) error {
	clientSubscriptions, ok := s.subscriptions[clientId]
	if !ok {
		clientSubscriptions = make(map[string]query.Query)
		s.subscriptions[clientId] = clientSubscriptions
	}
	for _, q := range newQueries {
		queryHash, err := query.StableStringify(q)
		if err != nil {
			return err
		}
		clientSubscriptions[queryHash] = q
	}
	return nil
}

// AddSubscription 为指定客户端添加一个查询订阅
// 参数:
//   - clientId: 客户端标识符
//   - q: 要添加的查询
//
// 返回:
//   - 如果查询哈希生成失败，返回错误
func (s *ActiveSet) AddSubscription(clientId string, q query.Query) error {
	queryHash, err := query.StableStringify(q)
	if err != nil {
		return err
	}

	if _, ok := s.subscriptions[clientId]; !ok {
		s.subscriptions[clientId] = make(map[string]query.Query)
	}
	s.subscriptions[clientId][queryHash] = q
	return nil
}

// RemoveSubscription 移除指定客户端的查询订阅
// 参数:
//   - clientId: 客户端标识符
//   - q: 要移除的查询
//
// 返回:
//   - 如果查询哈希生成失败，返回错误
func (s *ActiveSet) RemoveSubscription(clientId string, q query.Query) error {
	queryHash, err := query.StableStringify(q)
	if err != nil {
		return err
	}

	if clientMap, ok := s.subscriptions[clientId]; ok {
		delete(clientMap, queryHash)
	}
	return nil
}

// Contains 检查指定客户端是否订阅了给定的查询
// 参数:
//   - clientId: 客户端标识符
//   - q: 要检查的查询
//
// 返回:
//   - 布尔值表示是否包含该订阅
//   - 如果查询哈希生成失败，返回错误
func (a *ActiveSet) Contains(clientId string, q query.Query) (bool, error) {
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

// GetSubscriptions 获取指定客户端的所有查询订阅
// 参数:
//   - clientId: 客户端标识符
//
// 返回:
//   - 该客户端的所有查询订阅列表
func (a *ActiveSet) GetSubscriptions(clientId string) []query.Query {
	clientMap, ok := a.subscriptions[clientId]
	if !ok {
		return []query.Query{}
	}

	subscriptions := make([]query.Query, 0, len(clientMap))
	for _, q := range clientMap {
		subscriptions = append(subscriptions, q.(query.Query))
	}
	return subscriptions
}
