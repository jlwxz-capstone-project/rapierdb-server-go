package synchronizer

import (
	"bytes"
	"context"
	"errors"
	"slices"
	"sync"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

var (
	ErrInvalidStorageEvent = errors.New("invalid storage event")
)

// Synchronizer status constants
type SynchronizerStatus string

const (
	SynchronizerStatusStopped  SynchronizerStatus = "stopped"
	SynchronizerStatusStarting SynchronizerStatus = "starting"
	SynchronizerStatusRunning  SynchronizerStatus = "running"
	SynchronizerStatusStopping SynchronizerStatus = "stopping"
)

type SynchronizerConfig struct {
}

type Synchronizer struct {
	storageEngine       *storage_engine.StorageEngine
	storageEngineEvents *StorageEngineEvents
	queryExecutor       *query.QueryExecutor
	network             network_server.NetworkProvider
	config              SynchronizerConfig
	permission          *query.Permissions
	ctx                 context.Context
	cancel              context.CancelFunc
	queryManager        *QueryManager

	// Fields related to status
	status     SynchronizerStatus
	statusLock sync.RWMutex
	eventBus   *util.EventBus[SynchronizerStatus]
}

type StorageEngineEvents struct {
	CommittedCh  <-chan *storage_engine.TransactionCommittedEvent
	CanceledCh   <-chan *storage_engine.TransactionCanceledEvent
	RollbackedCh <-chan *storage_engine.TransactionRollbackedEvent
}

func NewSynchronizer(storageEngine *storage_engine.StorageEngine, channel network_server.NetworkProvider, config *SynchronizerConfig) *Synchronizer {
	return NewSynchronizerWithContext(context.Background(), storageEngine, channel, config)
}

func NewSynchronizerWithContext(ctx context.Context, storageEngine *storage_engine.StorageEngine, channel network_server.NetworkProvider, config *SynchronizerConfig) *Synchronizer {
	// Use default config if nil
	if config == nil {
		config = &SynchronizerConfig{}
	}

	log.Debugf("NewSynchronizer: Creating synchronizer")
	permission, err := query.NewPermissionFromJs(storageEngine.GetPermissionsJs())
	if err != nil {
		log.Errorf("NewSynchronizer: Failed to create permissions: %v", err)
	}

	// Create context
	ctx, cancel := context.WithCancel(ctx)

	queryExecutor := query.NewQueryExecutor(storageEngine)
	synchronizer := &Synchronizer{
		storageEngine:       storageEngine,
		storageEngineEvents: &StorageEngineEvents{},
		queryExecutor:       queryExecutor,
		network:             channel,
		config:              *config,
		permission:          permission,
		queryManager:        NewQueryManager(queryExecutor, permission),
		ctx:                 ctx,
		cancel:              cancel,
		status:              SynchronizerStatusStopped,
		statusLock:          sync.RWMutex{},
		eventBus:            util.NewEventBus[SynchronizerStatus](),
	}
	log.Debugf("NewSynchronizer: Synchronizer created successfully")
	return synchronizer
}

// GetStatus returns the current status of the synchronizer
func (s *Synchronizer) GetStatus() SynchronizerStatus {
	s.statusLock.RLock()
	defer s.statusLock.RUnlock()
	return s.status
}

// setStatus sets the synchronizer status (internal use)
func (s *Synchronizer) setStatus(status SynchronizerStatus) {
	s.statusLock.Lock()
	oldStatus := s.status
	s.status = status
	s.statusLock.Unlock()

	// Only publish event if status actually changed
	if oldStatus != status {
		s.eventBus.Publish(status)
	}
}

// SubscribeStatusChange subscribes to status change events
func (s *Synchronizer) SubscribeStatusChange() <-chan SynchronizerStatus {
	return s.eventBus.Subscribe()
}

// UnsubscribeStatusChange unsubscribes from status change events
func (s *Synchronizer) UnsubscribeStatusChange(ch <-chan SynchronizerStatus) {
	s.eventBus.Unsubscribe(ch)
}

// WaitForStatus waits for the synchronizer to reach the target status
func (s *Synchronizer) WaitForStatus(targetStatus SynchronizerStatus) <-chan struct{} {
	statusCh := s.SubscribeStatusChange()
	cleanup := func() {
		s.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(s.GetStatus, targetStatus, statusCh, cleanup, 0)
}

func (s *Synchronizer) Start() error {
	log.Debugf("Synchronizer.Start: Starting synchronizer")
	s.setStatus(SynchronizerStatusStarting)

	// Subscribe to storage engine events
	committedCh := s.storageEngine.SubscribeCommitted()
	canceledCh := s.storageEngine.SubscribeCanceled()
	rollbackedCh := s.storageEngine.SubscribeRollbacked()
	log.Debugf("Synchronizer.Start: Subscribed to storage engine events")

	// Save subscription channels for later cleanup
	s.storageEngineEvents = &StorageEngineEvents{
		CommittedCh:  committedCh,
		CanceledCh:   canceledCh,
		RollbackedCh: rollbackedCh,
	}

	go func() {
		log.Debugf("Synchronizer: Event handler goroutine started")
		for {
			select {
			case <-s.ctx.Done():
				log.Debugf("Synchronizer: Received cancel signal, event handler goroutine exiting")
				s.Stop()
				return
			case event := <-committedCh:
				log.Debugf("Synchronizer: Received transaction committed event")
				s.handleTransactionCommitted(event)
			case event := <-canceledCh:
				log.Debugf("Synchronizer: Received transaction canceled event")
				s.handleTransactionCanceled(event)
			case event := <-rollbackedCh:
				log.Debugf("Synchronizer: Received transaction rollbacked event")
				s.handleTransactionRollbacked(event)
			}
		}
	}()

	msgHandler := func(clientId string, msgBytes []byte) {
		log.Debugf("Synchronizer.msgHandler: Received message from client %s, length %d bytes", clientId, len(msgBytes))
		buf := bytes.NewBuffer(msgBytes)
		msg, err := message.DecodeMessage(buf)
		if err != nil {
			log.Errorf("msgHandler: Failed to decode message: %#+v", err)
			return
		}
		switch msg := msg.(type) {
		case *message.PostTransactionMessageV1:
			// Commit transaction to storage engine
			// Set committer ID for the transaction
			log.Debugf("msgHandler: Received %s from %s", msg.DebugSprint(), clientId)
			msg.Transaction.Committer = clientId
			log.Debugf("msgHandler: Committing transaction %s to storage engine", msg.Transaction.TxID)
			err = s.storageEngine.Commit(msg.Transaction)
			if err != nil {
				log.Errorf("msgHandler: Failed to commit transaction: %v", err)
				return
			}
			log.Debugf("msgHandler: Transaction %s committed successfully", msg.Transaction.TxID)

		case *message.SubscriptionUpdateMessageV1:
			log.Debugf("msgHandler: Received %s from %s", msg.DebugSprint(), clientId)
			// Handle removed subscriptions
			for _, q := range msg.Removed {
				log.Debugf("msgHandler: Client %s unsubscribed query %s", clientId, q.DebugSprint())
				err := s.queryManager.RemoveSubscriptedQuery(clientId, q)
				if err != nil {
					log.Errorf("msgHandler: Failed to remove subscription: %v", err)
				}
			}

			// Handle added subscriptions
			vqm := message.NewVersionQueryMessageV1()
			for _, q := range msg.Added {
				log.Debugf("msgHandler: Client %s subscribed query %s", clientId, q.DebugSprint())
				err := s.queryManager.SubscribeNewQuery(clientId, q)
				if err != nil {
					log.Errorf("msgHandler: Failed to add subscription: %v", err)
				}

				var queryCollection string
				switch q := q.(type) {
				case *query.FindManyQuery:
					queryCollection = q.Collection
				case *query.FindOneQuery:
					queryCollection = q.Collection
				}

				// TODO: Ideally, only documents included in the query result should be added to vqm.
				// For simplicity, all documents in the collection are added to vqm here.
				allCollections := vqm.GetAllCollections()
				if !slices.Contains(allCollections, queryCollection) {
					docs, err := s.storageEngine.LoadAllDocsInCollection(queryCollection, true)
					if err != nil {
						log.Errorf("msgHandler: Failed to get all documents in collection %s: %v", queryCollection, err)
						continue
					}
					for docId := range docs {
						vqm.AddDoc(queryCollection, docId)
					}
				}
			}

			vqmBytes, err := vqm.Encode()
			if err != nil {
				log.Errorf("msgHandler: Failed to encode version query message: %v", err)
				return
			}

			log.Debugf("msgHandler: Sending version query message %s to client %s", vqm.DebugSprint(), clientId)
			s.network.Send(clientId, vqmBytes)

		case *message.VersionQueryRespMessageV1:
			log.Debugf("msgHandler: Received %s from %s", msg.DebugSprint(), clientId)
			toUpsert := make(map[string][]byte)
			toDelete := make([]string, 0)
			for docKey, vvBytes := range msg.Responses {
				docKeyBytes := util.String2Bytes(docKey)
				collection := storage_engine.GetCollectionNameFromKey(docKeyBytes)
				docId := storage_engine.GetDocIdFromKey(docKeyBytes)
				if vvBytes == nil || len(vvBytes) == 0 {
					doc, err := s.storageEngine.LoadDoc(collection, docId)
					if err != nil {
						log.Errorf("msgHandler: Failed to load document %s/%s: %v", collection, docId, err)
						continue
					}
					docBytes := doc.ExportSnapshot().Bytes()
					toUpsert[docKey] = docBytes
				} else {
					doc, err := s.storageEngine.LoadDoc(collection, docId)
					if err != nil {
						log.Errorf("msgHandler: Failed to load document %s/%s: %v", collection, docId, err)
						continue
					}
					vv := loro.NewVvFromBytes(loro.NewRustBytesVec(vvBytes))
					updateBytesVec := doc.ExportUpdatesFrom(vv)
					updateBytes := updateBytesVec.Bytes()
					toUpsert[docKey] = updateBytes
				}
			}
			if len(toUpsert) > 0 {
				syncMsg := message.PostDocMessageV1{
					Upsert: toUpsert,
					Delete: toDelete,
				}
				log.Debugf("msgHandler: Sending sync message %s to client %s", syncMsg.DebugSprint(), clientId)
				syncMsgBytes, err := syncMsg.Encode()
				if err != nil {
					log.Errorf("msgHandler: Failed to encode sync message: %v", err)
					return
				}
				s.network.Send(clientId, syncMsgBytes)
			}
		}
	}
	s.network.SetMsgHandler(msgHandler)

	// Set status to running
	s.setStatus(SynchronizerStatusRunning)
	log.Debugf("Synchronizer.Start: Synchronizer started")

	return nil
}

// Stop stops the synchronizer
func (s *Synchronizer) Stop() {
	if s.status != SynchronizerStatusRunning {
		log.Warn("Synchronizer.Stop: Synchronizer is not running, no need to stop")
		return
	}

	log.Debugf("Synchronizer.Stop: Stopping synchronizer")
	log.Debugf("Synchronizer.Stop: Stopping synchronizer")
	s.setStatus(SynchronizerStatusStopping)

	// Unsubscribe from storage engine events
	s.storageEngine.UnsubscribeCommitted(s.storageEngineEvents.CommittedCh)
	s.storageEngine.UnsubscribeCanceled(s.storageEngineEvents.CanceledCh)
	s.storageEngine.UnsubscribeRollbacked(s.storageEngineEvents.RollbackedCh)
	log.Debugf("Synchronizer.Stop: Unsubscribed from storage engine events")

	// Close all client connections
	s.network.CloseAllConnections()

	// Cancel context, notify all goroutines to exit
	s.cancel()

	log.Debugf("Synchronizer.Stop: Synchronizer stopped")
	s.setStatus(SynchronizerStatusStopped)
}

// handleTransactionCommitted handles transaction committed events
// Note: This method catches all errors to ensure it doesn't break on error
func (s *Synchronizer) handleTransactionCommitted(event_ any) {
	log.Debugf("handleTransactionCommitted: Handling transaction committed event")
	event, ok := event_.(*storage_engine.TransactionCommittedEvent)
	if !ok {
		log.Errorf("handleTransactionCommitted: Event is not of type *storage.TransactionCommittedEvent")
		return
	}
	log.Debugf("handleTransactionCommitted: Handling transaction %s committed event, committer %s", event.Transaction.TxID, event.Committer)

	// Send AckTransactionMessage to the node that committed the transaction
	ackMsg := message.AckTransactionMessageV1{
		TxID: event.Transaction.TxID,
	}
	ackMsgBytes, err := ackMsg.Encode()
	if err != nil {
		log.Errorf("handleTransactionCommitted: Failed to encode ack message: %v", err)
	}
	log.Debugf("handleTransactionCommitted: Sending ack message to committer %s", event.Committer)
	err = s.network.Send(event.Committer, ackMsgBytes)
	if err != nil {
		log.Errorf("handleTransactionCommitted: Failed to send ack message: %v", err)
	}

	// Notify queryManager of the transaction
	// queryManager updates all query results based on the transaction
	// and returns the updates each client should see in cus
	cus := s.queryManager.HandleTransaction(event.Transaction)
	for clientId, cu := range cus {
		// Skip the committer client, since it already received the ack message
		if clientId == event.Committer {
			log.Debugf("handleTransactionCommitted: Skipping committer %s", clientId)
			continue
		}

		if cu.IsEmpty() {
			log.Debugf("handleTransactionCommitted: Client %s has no documents to sync", clientId)
		} else {
			log.Debugf("handleTransactionCommitted: Sending sync message to client %s, upsert: %d, delete: %d",
				clientId, len(cu.Updates), len(cu.Deletes))

			deletedKeys := make([]string, 0, len(cu.Deletes))
			for docKey := range cu.Deletes {
				deletedKeys = append(deletedKeys, docKey)
			}

			syncMsg := message.PostDocMessageV1{
				Upsert: cu.Updates,
				Delete: deletedKeys,
			}

			syncMsgBytes, err := syncMsg.Encode()
			if err != nil {
				log.Errorf("handleTransactionCommitted: Failed to encode sync message for client %s: %v",
					clientId, err)
				continue
			}

			err = s.network.Send(clientId, syncMsgBytes)
			if err != nil {
				log.Errorf("handleTransactionCommitted: Failed to send sync message to client %s: %v",
					clientId, err)
			} else {
				log.Debugf("handleTransactionCommitted: Successfully sent sync message to client %s", clientId)
			}
		}
	}

	log.Debugf("handleTransactionCommitted: Transaction %s handled", event.Transaction.TxID)
}

func (s *Synchronizer) handleTransactionCanceled(event any) {
	log.Debugf("handleTransactionCanceled: Handling transaction canceled event")
	// Send TransactionFailedMessage to the node whose transaction was canceled
	canceledEvent, ok := event.(*storage_engine.TransactionCanceledEvent)
	if !ok {
		log.Errorf("handleTransactionCanceled: Event is not of type *storage.TransactionCanceledEvent")
		return
	}
	log.Debugf("handleTransactionCanceled: Handling transaction %s canceled event, committer %s, reason: %s",
		canceledEvent.Transaction.TxID, canceledEvent.Committer, canceledEvent.Reason)

	failedMsg := message.TransactionFailedMessageV1{
		TxID:   canceledEvent.Transaction.TxID,
		Reason: canceledEvent.Reason,
	}

	failedMsgBytes, err := failedMsg.Encode()
	if err != nil {
		log.Errorf("handleTransactionCanceled: Failed to encode failed message: %v", err)
		return
	}

	log.Debugf("handleTransactionCanceled: Sending failed message to committer %s", canceledEvent.Committer)
	err = s.network.Send(canceledEvent.Committer, failedMsgBytes)
	if err != nil {
		log.Errorf("handleTransactionCanceled: Failed to send failed message: %v", err)
	} else {
		log.Debugf("handleTransactionCanceled: Successfully sent failed message")
	}
}

func (s *Synchronizer) handleTransactionRollbacked(event any) {
	log.Debugf("handleTransactionRollbacked: Handling transaction rollbacked event")
	// Send TransactionFailedMessage to the node whose transaction was rollbacked
	rollbackedEvent, ok := event.(*storage_engine.TransactionRollbackedEvent)
	if !ok {
		log.Errorf("handleTransactionRollbacked: Event is not of type *storage.TransactionRollbackedEvent")
		return
	}
	log.Debugf("handleTransactionRollbacked: Handling transaction %s rollbacked event, committer %s, reason: %s",
		rollbackedEvent.Transaction.TxID, rollbackedEvent.Committer, rollbackedEvent.Reason)

	failedMsg := message.TransactionFailedMessageV1{
		TxID:   rollbackedEvent.Transaction.TxID,
		Reason: rollbackedEvent.Reason,
	}

	failedMsgBytes, err := failedMsg.Encode()
	if err != nil {
		log.Errorf("handleTransactionRollbacked: Failed to encode failed message: %v", err)
		return
	}

	log.Debugf("handleTransactionRollbacked: Sending failed message to committer %s", rollbackedEvent.Committer)
	err = s.network.Send(rollbackedEvent.Committer, failedMsgBytes)
	if err != nil {
		log.Errorf("handleTransactionRollbacked: Failed to send failed message: %v", err)
	} else {
		log.Debugf("handleTransactionRollbacked: Successfully sent failed message")
	}
}
