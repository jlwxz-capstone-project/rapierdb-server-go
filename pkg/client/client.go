package client

import (
	"bytes"
	"context"
	"fmt"
	"sync"

	pe "github.com/pkg/errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	network_client "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/client"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type TestClientStatus int

const (
	TEST_CLIENT_STATUS_READY     TestClientStatus = 0
	TEST_CLIENT_STATUS_NOT_READY TestClientStatus = 1
)

// 客户端是否开启乐观更新
// 开启乐观更新：对每个客户端事务，立即修改客户端数据库，不等待服务端确认
// 不开启乐观更新：对每个客户端事务，不立即修改客户端数据库，等待服务端确认后才修改
type TestClientOptimisticMode int

const (
	TEST_CLIENT_OPTIMISTIC_MODE_ON  TestClientOptimisticMode = 0
	TEST_CLIENT_OPTIMISTIC_MODE_OFF TestClientOptimisticMode = 1
)

type TestClient struct {
	Docs           map[string]map[string]*loro.LoroDoc
	Queries        map[string]ReactiveQuery
	Channel        network_client.ClientChannel
	ChannelManager network_client.ClientChannelManager
	OptimisticMode TestClientOptimisticMode
	PendingQueue   map[string]*storage_engine.Transaction
	ctx            context.Context
	cancel         context.CancelFunc

	// 状态相关字段
	status     TestClientStatus
	statusLock sync.RWMutex
	statusEb   *util.EventBus[TestClientStatus]
}

func NewTestClient(channelManager network_client.ClientChannelManager) *TestClient {
	ctx, cancel := context.WithCancel(context.Background())
	client := &TestClient{
		Docs:           make(map[string]map[string]*loro.LoroDoc),
		Queries:        make(map[string]ReactiveQuery),
		Channel:        nil,
		ChannelManager: channelManager,
		OptimisticMode: TEST_CLIENT_OPTIMISTIC_MODE_OFF,
		PendingQueue:   make(map[string]*storage_engine.Transaction),
		ctx:            ctx,
		cancel:         cancel,

		status:     TEST_CLIENT_STATUS_NOT_READY,
		statusLock: sync.RWMutex{},
		statusEb:   util.NewEventBus[TestClientStatus](),
	}
	return client
}

func (c *TestClient) Connect() error {
	channel, err := c.ChannelManager.GetChannel()
	if err != nil {
		c.setStatus(TEST_CLIENT_STATUS_NOT_READY)
		return pe.WithStack(fmt.Errorf("获取通道失败: %v", err))
	}
	c.Channel = channel
	channel.SetMsgHandler(c.handleServerMessage)

	channelStatusCh := channel.SubscribeStatusChange()
	go func() {
		defer channel.UnsubscribeStatusChange(channelStatusCh)
		for status := range channelStatusCh {
			if status == network_client.CHANNEL_STATUS_READY {
				c.setStatus(TEST_CLIENT_STATUS_READY)
			} else {
				c.setStatus(TEST_CLIENT_STATUS_NOT_READY)
			}
		}
	}()

	// 将通道状态同步到客户端状态
	if channel.GetStatus() == network_client.CHANNEL_STATUS_READY {
		c.setStatus(TEST_CLIENT_STATUS_READY)
	} else {
		c.setStatus(TEST_CLIENT_STATUS_NOT_READY)
	}

	return nil
}

func (c *TestClient) Close() error {
	if c.Channel == nil {
		return pe.WithStack(fmt.Errorf("通道未连接"))
	}
	err := c.ChannelManager.Close(c.Channel)
	// 状态会由monitorChannelStatus自动更新，不需要在这里设置
	return err
}

func (c *TestClient) IsReady() bool {
	return c.GetStatus() == TEST_CLIENT_STATUS_READY
}

func (c *TestClient) handleServerMessage(data []byte) {
	buf := bytes.NewBuffer(data)
	msg, err := message.DecodeMessage(buf)
	if err != nil {
		log.Warnf("客户端解码消息失败: %v", err)
		return
	}

	log.Debugf("客户端收到 %s", msg.DebugSprint())

	switch msg := msg.(type) {
	case *message.AckTransactionMessageV1:
		if tx, ok := c.PendingQueue[msg.TxID]; ok {
			delete(c.PendingQueue, msg.TxID)
			// 如果未启用乐观更新，则收到事务确认消息后应用事务
			if c.OptimisticMode == TEST_CLIENT_OPTIMISTIC_MODE_OFF {
				c.applyTransaction(tx)
				log.Debugf("客户端收到事务确认消息，应用事务 %s", msg.TxID)
			}
		}
	case *message.VersionQueryMessageV1:
		resp := &message.VersionQueryRespMessageV1{
			ID:        msg.ID,
			Responses: make(map[string][]byte),
		}

		for docKey := range msg.Queries {
			collection := storage_engine.GetCollectionNameFromKey([]byte(docKey))
			docId := storage_engine.GetDocIdFromKey([]byte(docKey))

			var version []byte = nil
			if docs, ok := c.Docs[collection]; ok {
				if doc, ok := docs[docId]; ok {
					version = doc.GetStateVv().Encode().Bytes()
				}
			}

			if version == nil {
				resp.Responses[docKey] = []byte{}
			} else {
				resp.Responses[docKey] = version
			}
		}

		// 将响应发送回服务器
		respData, err := resp.Encode()
		if err != nil {
			log.Warnf("编码响应失败: %v", err)
			return
		}

		err = c.Channel.Send(respData)
		if err != nil {
			log.Warnf("发送响应失败: %v", err)
			return
		}
	case *message.PostDocMessageV1:
		for docKey, bytes := range msg.Upsert {
			collection := storage_engine.GetCollectionNameFromKey([]byte(docKey))
			docId := storage_engine.GetDocIdFromKey([]byte(docKey))

			var doc *loro.LoroDoc = nil
			if docs, ok := c.Docs[collection]; ok {
				if d, ok := docs[docId]; ok {
					doc = d
				}
			}

			if doc == nil {
				doc = loro.NewLoroDoc()
				doc.Import(bytes)
				log.Debugf("文档 %s.%s 不存在，加入到本地数据库", collection, docId)
				if docs, ok := c.Docs[collection]; ok {
					docs[docId] = doc
				} else {
					c.Docs[collection] = map[string]*loro.LoroDoc{
						docId: doc,
					}
				}
			} else {
				log.Debugf("文档 %s.%s 存在，更新本地数据库", collection, docId)
				doc.Import(bytes)
			}
		}

		// TODO
		for _, query := range c.Queries {
			c.updateQueryResult(query)
		}
	}
}

func (c *TestClient) FindOne(q *query.FindOneQuery) (*ReactiveFindOneQuery, error) {
	queryHash, err := query.StableStringify(q)
	if err != nil {
		return nil, err
	}

	if q, ok := c.Queries[queryHash]; ok {
		if q, ok := q.(*ReactiveFindOneQuery); ok {
			return q, nil
		} else {
			return nil, pe.WithStack(fmt.Errorf("query %s is not a FindOneQuery", q.Query.DebugSprint()))
		}
	}

	// 向服务端更新订阅
	go func() {
		added := []query.Query{q}
		removed := []query.Query{}
		c.updateSubscription(added, removed)
	}()

	reactiveQuery := &ReactiveFindOneQuery{
		Query:  q,
		Status: NewRef(QUERY_STATUS_LOADING),
		Error:  NewRef[error](nil),
		Result: NewRef[*DocWithId](nil),
	}
	c.Queries[queryHash] = reactiveQuery
	c.updateQueryResult(reactiveQuery)
	return reactiveQuery, nil
}

func (c *TestClient) FindMany(q *query.FindManyQuery) (*ReactiveFindManyQuery, error) {
	queryHash, err := query.StableStringify(q)
	if err != nil {
		return nil, err
	}

	if q, ok := c.Queries[queryHash]; ok {
		if q, ok := q.(*ReactiveFindManyQuery); ok {
			return q, nil
		} else {
			return nil, pe.WithStack(fmt.Errorf("query %s is not a FindManyQuery", q.Query.DebugSprint()))
		}
	}

	// 向服务端更新订阅
	go func() {
		added := []query.Query{q}
		removed := []query.Query{}
		c.updateSubscription(added, removed)
	}()

	reactiveQuery := &ReactiveFindManyQuery{
		Query:  q,
		Status: NewRef(QUERY_STATUS_LOADING),
		Error:  NewRef[error](nil),
		Result: NewRef([]*DocWithId{}),
	}
	c.Queries[queryHash] = reactiveQuery
	c.updateQueryResult(reactiveQuery)
	return reactiveQuery, nil
}

func (c *TestClient) SubmitTransaction(tx *storage_engine.Transaction) error {
	log.Debugf("客户端提交事务 %s", tx.TxID)
	// 如果启用乐观更新，则立即修改客户端数据库
	if c.OptimisticMode == TEST_CLIENT_OPTIMISTIC_MODE_ON {
		log.Debugf("客户端启用乐观更新，立即更新客户端数据库 %s", tx.TxID)
		c.applyTransaction(tx)
	}

	// 将事务加入到待确认队列
	if c.PendingQueue[tx.TxID] != nil {
		return pe.WithStack(fmt.Errorf("事务 %s 已存在", tx.TxID))
	}
	log.Debugf("客户端将事务 %s 加入到待确认队列", tx.TxID)
	c.PendingQueue[tx.TxID] = tx

	// 向服务端发送 PostTransactionMessage 消息
	msg := message.PostTransactionMessageV1{
		Transaction: tx,
	}
	msgBytes, err := msg.Encode()
	if err != nil {
		log.Warnf("编码 PostTransactionMessage 消息失败: %v", err)
		return nil
	}
	err = c.Channel.Send(msgBytes)
	if err != nil {
		log.Warnf("发送 PostTransactionMessage 消息失败: %v", err)
		return nil
	}
	return nil
}

func (c *TestClient) applyTransaction(tx *storage_engine.Transaction) error {
	for _, op := range tx.Operations {
		switch op := op.(type) {
		case *storage_engine.InsertOp:
			if _, exists := c.Docs[op.Collection]; !exists {
				c.Docs[op.Collection] = make(map[string]*loro.LoroDoc)
			}
			doc := loro.NewLoroDoc()
			doc.Import(op.Snapshot)
			c.Docs[op.Collection][op.DocID] = doc
		case *storage_engine.UpdateOp:
			if _, exists := c.Docs[op.Collection]; !exists {
				c.Docs[op.Collection] = make(map[string]*loro.LoroDoc)
			}
			doc := loro.NewLoroDoc()
			doc.Import(op.Update)
			c.Docs[op.Collection][op.DocID] = doc
		case *storage_engine.DeleteOp:
			if _, exists := c.Docs[op.Collection]; exists {
				delete(c.Docs[op.Collection], op.DocID)
				if len(c.Docs[op.Collection]) == 0 {
					delete(c.Docs, op.Collection)
				}
			}
		}
	}

	// TODO
	for _, query := range c.Queries {
		c.updateQueryResult(query)
	}

	return nil
}

func (c *TestClient) updateSubscription(added []query.Query, removed []query.Query) {
	msg := &message.SubscriptionUpdateMessageV1{
		Added:   added,
		Removed: removed,
	}
	msgBytes, err := msg.Encode()
	if err != nil {
		log.Warnf("编码订阅更新消息失败: %v", err)
		return
	}
	err = c.Channel.Send(msgBytes)
	if err != nil {
		log.Warnf("发送订阅更新消息失败: %v", err)
	}
}

func (c *TestClient) updateQueryResult(q ReactiveQuery) {
	switch q := q.(type) {
	case *ReactiveFindOneQuery:
		c.updateFindOneQueryResult(q)
	case *ReactiveFindManyQuery:
		c.updateFindManyQueryResult(q)
	}
}

func (c *TestClient) updateFindOneQueryResult(q *ReactiveFindOneQuery) {
	docs, ok := c.Docs[q.Query.Collection]
	if !ok {
		log.Warnf("collection %s not found", q.Query.Collection)
		return
	}

	for docId, doc := range docs {
		matched, err := q.Query.Match(doc)
		if err != nil {
			log.Warnf("error matching document %s: %v", docId, err)
			return
		}
		if matched {
			q.Result.Set(&DocWithId{
				DocId: docId,
				Doc:   doc,
			})
			return
		}
	}
}

func (c *TestClient) updateFindManyQueryResult(q *ReactiveFindManyQuery) {
	docs, ok := c.Docs[q.Query.Collection]
	if !ok {
		log.Warnf("collection %s not found", q.Query.Collection)
		return
	}

	results := make([]*DocWithId, 0)
	for docId, doc := range docs {
		matched, err := q.Query.Match(doc)
		if err != nil {
			log.Warnf("error matching document %s: %v", docId, err)
			continue
		}
		if matched {
			results = append(results, &DocWithId{
				DocId: docId,
				Doc:   doc,
			})
		}
	}
	q.Result.Set(results)
}

// GetStatus 获取客户端状态
func (c *TestClient) GetStatus() TestClientStatus {
	c.statusLock.RLock()
	defer c.statusLock.RUnlock()
	return c.status
}

// setStatus 设置客户端状态
func (c *TestClient) setStatus(status TestClientStatus) {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	oldStatus := c.status
	c.status = status

	if oldStatus != status {
		c.statusEb.Publish(status)
	}
}

// SubscribeStatusChange 订阅状态变更事件
func (c *TestClient) SubscribeStatusChange() <-chan TestClientStatus {
	return c.statusEb.Subscribe()
}

// UnsubscribeStatusChange 取消订阅状态变更事件
func (c *TestClient) UnsubscribeStatusChange(ch <-chan TestClientStatus) {
	c.statusEb.Unsubscribe(ch)
}

// WaitForStatus 等待客户端达到指定状态
func (c *TestClient) WaitForStatus(targetStatus TestClientStatus) <-chan struct{} {
	statusCh := c.SubscribeStatusChange()
	cleanup := func() {
		c.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(c.GetStatus, targetStatus, statusCh, cleanup, 0)
}
