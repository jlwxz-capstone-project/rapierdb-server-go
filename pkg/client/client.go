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

func (s TestClientStatus) String() string {
	switch s {
	case TEST_CLIENT_STATUS_READY:
		return "ready"
	case TEST_CLIENT_STATUS_NOT_READY:
		return "not ready"
	default:
		return fmt.Sprintf("unknown status: %d", s)
	}
}

// Whether the client enables optimistic updates
// If enabled: for each client transaction, immediately update the client database without waiting for server confirmation
type TestClientOptimisticMode int

const (
	TEST_CLIENT_OPTIMISTIC_MODE_ON  TestClientOptimisticMode = 0
	TEST_CLIENT_OPTIMISTIC_MODE_OFF TestClientOptimisticMode = 1
)

type TestClient struct {
	Docs            map[string]map[string]*loro.LoroDoc
	Queries         map[string]ReactiveQuery
	NetworkProvider network_client.NetworkProvider
	OptimisticMode  TestClientOptimisticMode
	PendingQueue    map[string]*storage_engine.Transaction
	ctx             context.Context
	cancel          context.CancelFunc

	// Status-related fields
	status     TestClientStatus
	statusLock sync.RWMutex
	statusEb   *util.EventBus[TestClientStatus]
}

func NewTestClient(networkProvider network_client.NetworkProvider) *TestClient {
	ctx, cancel := context.WithCancel(context.Background())
	client := &TestClient{
		Docs:            make(map[string]map[string]*loro.LoroDoc),
		Queries:         make(map[string]ReactiveQuery),
		NetworkProvider: networkProvider,
		OptimisticMode:  TEST_CLIENT_OPTIMISTIC_MODE_OFF,
		PendingQueue:    make(map[string]*storage_engine.Transaction),
		ctx:             ctx,
		cancel:          cancel,

		status:     TEST_CLIENT_STATUS_NOT_READY,
		statusLock: sync.RWMutex{},
		statusEb:   util.NewEventBus[TestClientStatus](),
	}
	return client
}

func (c *TestClient) Connect() error {
	// Check if the client is already connected or trying to connect
	if c.GetStatus() == TEST_CLIENT_STATUS_READY {
		log.Warnf("TestClient is already connected or connecting.")
		return nil // Or return an error if preferred
	}

	// Subscribe to status changes of the new NetworkProvider
	channelStatusCh := c.NetworkProvider.SubscribeStatusChange()
	go func() {
		defer c.NetworkProvider.UnsubscribeStatusChange(channelStatusCh)
		for status := range channelStatusCh {
			// Map NetworkStatus to TestClientStatus
			if status == network_client.NetworkReady {
				c.setStatus(TEST_CLIENT_STATUS_READY)
			} else {
				// Handle other network statuses appropriately
				// For now, any non-Ready status maps to NOT_READY
				c.setStatus(TEST_CLIENT_STATUS_NOT_READY)
			}
		}
		// When the channel closes (provider stopped/closed), ensure status is NOT_READY
		c.setStatus(TEST_CLIENT_STATUS_NOT_READY)
	}()

	err := c.NetworkProvider.Connect()
	if err != nil {
		// Do not set status here, let the subscription handle it
		return pe.WithStack(fmt.Errorf("Failed to connect network provider: %v", err))
	}

	// Set the message handler on the new NetworkProvider
	c.NetworkProvider.SetMsgHandler(c.handleServerMessage)

	return nil
}

func (c *TestClient) Close() error {
	if c.NetworkProvider == nil {
		return pe.WithStack(fmt.Errorf("Channel not connected"))
	}
	err := c.NetworkProvider.Close()
	// Status will be updated automatically by monitorChannelStatus, no need to set here
	return err
}

func (c *TestClient) IsReady() bool {
	return c.GetStatus() == TEST_CLIENT_STATUS_READY
}

func (c *TestClient) handleServerMessage(data []byte) {
	buf := bytes.NewBuffer(data)
	msg, err := message.DecodeMessage(buf)
	if err != nil {
		log.Warnf("Client failed to decode message: %v", err)
		return
	}

	log.Debugf("Client received %s", msg.DebugSprint())

	switch msg := msg.(type) {
	case *message.AckTransactionMessageV1:
		if tx, ok := c.PendingQueue[msg.TxID]; ok {
			delete(c.PendingQueue, msg.TxID)
			// If optimistic update is not enabled, apply the transaction after receiving confirmation
			if c.OptimisticMode == TEST_CLIENT_OPTIMISTIC_MODE_OFF {
				c.applyTransaction(tx)
				log.Debugf("Client received transaction confirmation, applied transaction %s", msg.TxID)
			}
		}
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

		// Send the response back to the server
		respData, err := resp.Encode()
		if err != nil {
			log.Warnf("Failed to encode response: %v", err)
			return
		}

		err = c.NetworkProvider.Send(respData)
		if err != nil {
			log.Warnf("Failed to send response: %v", err)
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
				log.Debugf("Document %s.%s does not exist, adding to local database", collection, docId)
				if docs, ok := c.Docs[collection]; ok {
					docs[docId] = doc
				} else {
					c.Docs[collection] = map[string]*loro.LoroDoc{
						docId: doc,
					}
				}
			} else {
				log.Debugf("Document %s.%s exists, updating local database", collection, docId)
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

	// Update subscription to the server
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

	// Update subscription to the server
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
	log.Debugf("Client submitting transaction %s", tx.TxID)
	// If optimistic update is enabled, immediately update the client database
	if c.OptimisticMode == TEST_CLIENT_OPTIMISTIC_MODE_ON {
		log.Debugf("Client optimistic update enabled, immediately updating client database %s", tx.TxID)
		c.applyTransaction(tx)
	}

	// Add the transaction to the pending confirmation queue
	if c.PendingQueue[tx.TxID] != nil {
		return pe.WithStack(fmt.Errorf("Transaction %s already exists", tx.TxID))
	}
	log.Debugf("Client adding transaction %s to pending queue", tx.TxID)
	c.PendingQueue[tx.TxID] = tx

	// Send PostTransactionMessage to the server
	msg := message.PostTransactionMessageV1{
		Transaction: tx,
	}
	msgBytes, err := msg.Encode()
	if err != nil {
		log.Warnf("Failed to encode PostTransactionMessage: %v", err)
		return nil
	}
	err = c.NetworkProvider.Send(msgBytes)
	if err != nil {
		log.Warnf("Failed to send PostTransactionMessage: %v", err)
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
		log.Warnf("Failed to encode subscription update message: %v", err)
		return
	}
	err = c.NetworkProvider.Send(msgBytes)
	if err != nil {
		log.Warnf("Failed to send subscription update message: %v", err)
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

// GetStatus returns the client status
func (c *TestClient) GetStatus() TestClientStatus {
	c.statusLock.RLock()
	defer c.statusLock.RUnlock()
	return c.status
}

// setStatus sets the client status
func (c *TestClient) setStatus(status TestClientStatus) {
	c.statusLock.Lock()
	defer c.statusLock.Unlock()
	oldStatus := c.status
	c.status = status

	if oldStatus != status {
		log.Debugf("Client status changed: %s -> %s", oldStatus, status)
		c.statusEb.Publish(status)
	}
}

// SubscribeStatusChange subscribes to status change events
func (c *TestClient) SubscribeStatusChange() <-chan TestClientStatus {
	return c.statusEb.Subscribe()
}

// UnsubscribeStatusChange unsubscribes from status change events
func (c *TestClient) UnsubscribeStatusChange(ch <-chan TestClientStatus) {
	c.statusEb.Unsubscribe(ch)
}

// WaitForStatus waits for the client to reach the specified status
func (c *TestClient) WaitForStatus(targetStatus TestClientStatus) <-chan struct{} {
	statusCh := c.SubscribeStatusChange()
	cleanup := func() {
		c.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(c.GetStatus, targetStatus, statusCh, cleanup, 0)
}
