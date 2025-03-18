package client

import (
	"bytes"
	"context"
	"fmt"
	"sync"
	"time"

	pe "github.com/pkg/errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	network_client "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/client"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type TestClientStatus int

const (
	TEST_CLIENT_STATUS_NOT_STARTED TestClientStatus = 0
	TEST_CLIENT_STATUS_CONNECTING  TestClientStatus = 1
	TEST_CLIENT_STATUS_CONNECTED   TestClientStatus = 2
	TEST_CLIENT_STATUS_CLOSED      TestClientStatus = 3
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

		status:     TEST_CLIENT_STATUS_NOT_STARTED,
		statusLock: sync.RWMutex{},
		statusEb:   util.NewEventBus[TestClientStatus](),
	}
	return client
}

func (c *TestClient) Connect() error {
	c.setStatus(TEST_CLIENT_STATUS_CONNECTING)
	channel, err := c.ChannelManager.GetChannel()
	if err != nil {
		c.setStatus(TEST_CLIENT_STATUS_NOT_STARTED)
		return pe.WithStack(fmt.Errorf("获取通道失败: %v", err))
	}
	c.Channel = channel
	channel.SetMsgHandler(c.handleServerMessage)

	// 将通道状态同步到客户端状态

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

func (c *TestClient) IsConnected() bool {
	return c.GetStatus() == TEST_CLIENT_STATUS_CONNECTED
}

func (c *TestClient) handleServerMessage(data []byte) {
	buf := bytes.NewBuffer(data)
	msg, err := message.DecodeMessage(buf)
	if err != nil {
		log.Warnf("解码消息失败: %v", err)
		return
	}

	switch msg := msg.(type) {
	case *message.VersionQueryMessageV1:
		resp := &message.VersionQueryRespMessageV1{
			Responses: make(map[string]map[string][]byte),
		}
		for collection, docIds := range msg.Queries {
			docs, ok := c.Docs[collection]
			cret := make(map[string][]byte)

			if !ok {
				// 如果集合不存在，则所有文档都不存在，用 []byte{} 表示不存在的文档
				for docId := range docIds {
					cret[docId] = []byte{}
				}
			} else {
				for docId := range docIds {
					doc, ok := docs[docId]
					if !ok {
						cret[docId] = []byte{} // 文档不存在
					} else {
						version := doc.GetStateVv().Encode().Bytes()
						cret[docId] = version
					}
				}
			}

			resp.Responses[collection] = cret
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
	}
	// 处理其他消息类型...
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

	reactiveQuery := &ReactiveFindOneQuery{
		Query:  q,
		Status: NewRef(QUERY_STATUS_LOADING),
		Error:  NewRef[error](nil),
		Result: NewRef[*DocWithId](nil),
	}
	c.Queries[queryHash] = reactiveQuery
	c.loadQueryResult(reactiveQuery)
	return reactiveQuery, nil
}

func (c *TestClient) FindMany(q *query.FindManyQuery) (ReactiveQuery, error) {
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

	reactiveQuery := &ReactiveFindManyQuery{
		Query:  q,
		Status: NewRef(QUERY_STATUS_LOADING),
		Error:  NewRef[error](nil),
		Result: NewRef([]*DocWithId{}),
	}
	c.Queries[queryHash] = reactiveQuery
	c.loadQueryResult(reactiveQuery)
	return reactiveQuery, nil
}

func (c *TestClient) loadQueryResult(q ReactiveQuery) {
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

// WaitForStatus 等待客户端达到指定状态，返回一个通道，当达到目标状态时会收到通知
// timeout 为等待超时时间，如果为0则永不超时
func (c *TestClient) WaitForStatus(targetStatus TestClientStatus, timeout time.Duration) <-chan struct{} {
	statusCh := c.SubscribeStatusChange()
	cleanup := func() {
		c.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(c.GetStatus, targetStatus, statusCh, cleanup, timeout)
}
