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

type TestClient struct {
	Docs           map[string]map[string]*loro.LoroDoc
	Queries        map[string]ReactiveQuery
	Channel        network_client.ClientChannel
	ChannelManager network_client.ClientChannelManager
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

	// 将通道状态同步到客户端状态
	if channel.GetStatus() == network_client.CHANNEL_STATUS_READY {
		c.setStatus(TEST_CLIENT_STATUS_READY)
	} else {
		c.setStatus(TEST_CLIENT_STATUS_NOT_READY)
	}

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

	log.Debugf("客户端收到 %s", msg.DebugPrint())

	switch msg := msg.(type) {
	case *message.VersionQueryMessageV1:
		resp := &message.VersionQueryRespMessageV1{
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
			return nil, pe.WithStack(fmt.Errorf("query %s is not a FindOneQuery", q.Query.DebugPrint()))
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
			return nil, pe.WithStack(fmt.Errorf("query %s is not a FindManyQuery", q.Query.DebugPrint()))
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
