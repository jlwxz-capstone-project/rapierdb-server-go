package subscription

import "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage"

type QuerySubscriptionManager struct {
	storageEngine *storage.StorageEngine
	// ClientId -> Subscriptions
	// 简单起见，现在订阅的粒度为集合，即要么订阅整个集合，要么不订阅
	// 注意：虽然现在订阅的粒度为集合，但集合中不符合权限的文档不会被返回
	subscriptions map[string][]string
}

func NewQuerySubscriptionManager(storageEngine *storage.StorageEngine) *QuerySubscriptionManager {
	return &QuerySubscriptionManager{
		storageEngine: storageEngine,
		subscriptions: make(map[string][]string),
	}
}

func (mgr *QuerySubscriptionManager) Subscribe(clientId string, collection string) {
	mgr.subscriptions[clientId] = append(mgr.subscriptions[clientId], collection)
}

func (mgr *QuerySubscriptionManager) Unsubscribe(clientId string, collection string) {
	subs := mgr.subscriptions[clientId]
	for i, sub := range subs {
		if sub == collection {
			// 从切片中删除该元素
			subs[i] = subs[len(subs)-1]
			mgr.subscriptions[clientId] = subs[:len(subs)-1]
			return
		}
	}
}

func (mgr *QuerySubscriptionManager) CalcDelta(e storage.TransactionCommittedEvent) {

}
