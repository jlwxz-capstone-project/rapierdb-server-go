package synchronizer2

import (
	"bytes"
	"context"
	"fmt"
	"sync/atomic"

	pe "github.com/pkg/errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_connector"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/key_utils"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/message/v1"
	network_server "github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/network/server"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/permission_proxy"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query_executor"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type SynchronizerStatus int32

const (
	SynchronizerStatusNotStarted SynchronizerStatus = 0
	SynchronizerStatusStarting   SynchronizerStatus = 1
	SynchronizerStatusRunning    SynchronizerStatus = 2
	SynchronizerStatusStopping   SynchronizerStatus = 3
	SynchronizerStatusStopped    SynchronizerStatus = 4
)

type Synchronizer struct {
	// Dependencies should be injected
	dbConnector db_connector.DbConnector
	network     network_server.NetworkProvider
	dbUrl       string

	// Managed databases
	// db url -> managed db (db connection, query executor, permission proxy)
	managedDb *ManagedDb

	// Context, used to stop the synchronizer
	ctx    context.Context
	cancel context.CancelFunc

	// Status
	status         atomic.Int32
	statusEventBus *util.EventBus[SynchronizerStatus]
}

type SynchronizerParams struct {
	DbConnector db_connector.DbConnector
	Network     network_server.NetworkProvider
	DbUrl       string
}

func NewSynchronizerWithContext(ctx context.Context, params *SynchronizerParams) *Synchronizer {
	ctx, cancel := context.WithCancel(ctx)

	synchronizer := &Synchronizer{
		dbConnector: params.DbConnector,
		network:     params.Network,
		dbUrl:       params.DbUrl,
		managedDb:   nil,
		ctx:         ctx,
		cancel:      cancel,
		// status init to SynchronizerStatusNotStarted by default
		statusEventBus: util.NewEventBus[SynchronizerStatus](),
	}

	return synchronizer
}

// Start starts the synchronizer
//
// block until the synchronizer is running
func (s *Synchronizer) Start() error {
	if !s.swapStatus(SynchronizerStatusNotStarted, SynchronizerStatusStarting) {
		return pe.Errorf("cannot start synchronizer, expect status SynchronizerStatusNotStarted, but got %s", s.GetStatus())
	}

	log.Debugf("Synchronizer.Start: Synchronizer starting")

	// connect to the database
	err := s.connectDatabase()
	if err != nil {
		return pe.Errorf("failed to connect to database: %v", err)
	}

	// if network is not running, wait for it
	if s.network.GetStatus() != network_server.NetworkRunning {
		log.Debugf("Synchronizer.Start: Network is not running, waiting for it")
		<-s.network.WaitForStatus(network_server.NetworkRunning)
		log.Debugf("Synchronizer.Start: Network is running now!")
	}

	// start a goroutine to stop the synchronizer when the context is done
	go func() {
		<-s.ctx.Done()
		log.Debugf("Synchronizer.Start: Detect context done, stopping synchronizer")
		s.Stop()
	}()

	s.network.SetMsgHandler(s.handleMessage)

	// start a goroutine to handle connection closed events
	connClosedCh := s.network.SubscribeConnectionClosed()
	go func() {
		defer s.network.UnsubscribeConnectionClosed(connClosedCh)
		select {
		case <-s.ctx.Done():
			return
		case ev := <-connClosedCh:
			s.handleConnectionClosed(ev)
		}
	}()

	if !s.swapStatus(SynchronizerStatusStarting, SynchronizerStatusRunning) {
		return pe.Errorf("cannot start synchronizer, expect status SynchronizerStatusStarting, but got %s", s.GetStatus())
	}

	log.Debugf("Synchronizer.Start: Synchronizer started")
	return nil
}

// Stop stops the synchronizer
//
// block until the synchronizer and all sub modules (network and managed db) are stopped
func (s *Synchronizer) Stop() error {
	if !s.swapStatus(SynchronizerStatusRunning, SynchronizerStatusStopping) {
		return pe.Errorf("cannot stop synchronizer, expect status SynchronizerStatusRunning, but got %s", s.GetStatus())
	}

	log.Debugf("Synchronizer.Stop: Synchronizer stopping")

	s.cancel()

	// wait for the database to be closed
	<-s.managedDb.conn.WaitForStatus(db_conn.DbConnStatusClosed)
	<-s.network.WaitForStatus(network_server.NetworkStopped)

	if !s.swapStatus(SynchronizerStatusStopping, SynchronizerStatusStopped) {
		return pe.Errorf("cannot stop synchronizer, expect status SynchronizerStatusStopping, but got %s", s.GetStatus())
	}

	log.Debugf("Synchronizer.Stop: Synchronizer stopped")
	return nil
}

// GetStatus returns the current status of the synchronizer
func (s *Synchronizer) GetStatus() SynchronizerStatus {
	return SynchronizerStatus(s.status.Load())
}

// swapStatus swaps the status of the synchronizer
// returns true if the status is swapped, false otherwise
func (s *Synchronizer) swapStatus(from, to SynchronizerStatus) bool {
	if from == to {
		return false
	}

	if s.status.CompareAndSwap(int32(from), int32(to)) {
		s.statusEventBus.Publish(to)
		return true
	}
	return false
}

// setStatus sets the status of the synchronizer
// it will not publish the status change event if the status is the same
func (s *Synchronizer) setStatus(status SynchronizerStatus) {
	currStatus := s.GetStatus()
	if currStatus == status {
		return
	}

	s.statusEventBus.Publish(status)
	s.status.Store(int32(status))
}

// SubscribeStatusChange subscribes to status change events
func (s *Synchronizer) SubscribeStatusChange() <-chan SynchronizerStatus {
	return s.statusEventBus.Subscribe()
}

// UnsubscribeStatusChange unsubscribes from status change events
func (s *Synchronizer) UnsubscribeStatusChange(ch <-chan SynchronizerStatus) {
	s.statusEventBus.Unsubscribe(ch)
}

func (s *Synchronizer) WaitForStatus(status SynchronizerStatus) <-chan struct{} {
	statusCh := s.SubscribeStatusChange()
	cleanup := func() {
		s.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(s.GetStatus, status, statusCh, cleanup, 0)
}

func (s *Synchronizer) handleMessage(clientId string, msgBytes []byte) {
	log.Debugf("Synchronizer.handleMessage: Received message from client %s, length %d bytes", clientId, len(msgBytes))

	// decode message
	buf := bytes.NewBuffer(msgBytes)
	msg, err := message.DecodeMessage(buf)
	if err != nil {
		log.Errorf("Synchronizer.handleMessage: Failed to decode message: %v, ignore it", err)
		return
	}
	log.Debugf("%#v", msg)
	log.Debugf("Synchronizer.handleMessage: Received %s from %s", msg.DebugSprint(), clientId)

	switch msg := msg.(type) {
	case *message.PostTransactionMessageV1:
		// set committer to client id
		msg.Transaction.Committer = clientId

		// authorization
		pass := true
		dbWrapper := &permission_proxy.DbWrapper{
			QueryExecutor: s.managedDb.queryExecutor,
		}
		for _, op := range msg.Transaction.Operations {
			switch op := op.(type) {
			case *db_conn.InsertOp:
				{
					newDoc := loro.NewLoroDoc()
					newDoc.Import(op.Snapshot)
					pass = s.managedDb.permissionProxy.CanCreate(permission_proxy.CanCreateParams{
						Collection: op.Collection,
						DocId:      op.DocID,
						NewDoc:     newDoc,
						ClientId:   clientId,
						Db:         dbWrapper,
					})
					if !pass {
						pass = false
						break
					}
				}
			case *db_conn.UpdateOp:
				{
					oldDoc, err := s.managedDb.conn.LoadDoc(op.Collection, op.DocID)
					if err != nil {
						pass = false
						log.Debugf("Synchronizer.handleMessage: trying to update doc %s.%s, but failed to load doc: %v", op.Collection, op.DocID, err)
						break
					}
					newDoc := oldDoc.Fork()
					newDoc.Import(op.Update)
					pass = s.managedDb.permissionProxy.CanUpdate(permission_proxy.CanUpdateParams{
						Collection: op.Collection,
						DocId:      op.DocID,
						NewDoc:     newDoc,
						OldDoc:     oldDoc,
						ClientId:   clientId,
						Db:         dbWrapper,
					})
					if !pass {
						pass = false
						break
					}
				}
			case *db_conn.DeleteOp:
				{
					oldDoc, err := s.managedDb.conn.LoadDoc(op.Collection, op.DocID)
					if err != nil {
						pass = false
						log.Debugf("Synchronizer.handleMessage: trying to delete doc %s.%s, but failed to load doc: %v", op.Collection, op.DocID, err)
						break
					}
					pass = s.managedDb.permissionProxy.CanDelete(permission_proxy.CanDeleteParams{
						Collection: op.Collection,
						DocId:      op.DocID,
						Doc:        oldDoc,
						ClientId:   clientId,
						Db:         dbWrapper,
					})
					if !pass {
						pass = false
						break
					}
				}
			}
		}
		if !pass {
			log.Errorf("Synchronizer.handleMessage: Transaction failed to pass authorization, ignore it")
			err := sendTransactionFailedMessage(
				s.network,
				clientId,
				msg.Transaction.TxID,
				fmt.Errorf("transaction failed to pass authorization"),
			)
			if err != nil {
				log.Errorf("Synchronizer.handleMessage: Failed to send transaction failed message to client %s: %v", clientId, err)
			}
			return
		}

		// commit transaction to storage engine
		// because we listen to transaction committed / rollbacked events
		// so we don't need to send TransactionAckMessage or
		// TransactionFailedMessage to client here
		s.managedDb.conn.Commit(msg.Transaction)
		log.Debugf("Synchronizer.handleMessage: Committed transaction %s", msg.Transaction.TxID)
		return

	case *message.SubscriptionUpdateMessageV1:
		// handle removed subscriptions
		for _, q := range msg.Removed {
			log.Debugf("Synchronizer.handleMessage: Client %s unsubscribed %s", clientId, q.DebugSprint())
			err := s.managedDb.queryManager.RemoveSubscriptedQuery(clientId, q)
			if err != nil {
				log.Errorf("Synchronizer.handleMessage: Failed to remove subscripted query %s: %v", q.DebugSprint(), err)
			}
		}

		// handle added subscriptions
		docKeys := map[string]struct{}{}
		for _, q := range msg.Added {
			log.Debugf("Synchronizer.handleMessage: Client %s subscribed %s", clientId, q.DebugSprint())
			err := s.managedDb.queryManager.SubscribeNewQuery(clientId, q)
			if err != nil {
				log.Errorf("Synchronizer.handleMessage: Failed to subscribe new query %s: %v", q.DebugSprint(), err)
				continue
			}

			// generate VersionQueryMessageV1
			// first exec the query, then collect all doc keys in query result
			switch q := q.(type) {
			case *query.FindOneQuery:
				res, err := s.managedDb.queryExecutor.FindOne(q)
				if err != nil {
					log.Errorf("Synchronizer.handleMessage: Failed to exec find one query %s: %v", q.DebugSprint(), err)
					continue
				}
				if res != nil {
					docKey, err := key_utils.CalcDocKey(q.Collection, res.DocId)
					if err != nil {
						log.Errorf("Synchronizer.handleMessage: Failed to calc doc key for find one query %s: %v", q.DebugSprint(), err)
						continue
					}
					docKeys[string(docKey)] = struct{}{}
				}
			case *query.FindManyQuery:
				res, err := s.managedDb.queryExecutor.FindMany(q)
				if err != nil {
					log.Errorf("Synchronizer.handleMessage: Failed to exec find many query %s: %v", q.DebugSprint(), err)
					continue
				}
				for _, docWithId := range res {
					docKey, err := key_utils.CalcDocKey(q.Collection, docWithId.DocId)
					if err != nil {
						log.Errorf("Synchronizer.handleMessage: Failed to calc doc key for find many query %s: %v", q.DebugSprint(), err)
						continue
					}
					docKeys[string(docKey)] = struct{}{}
				}
			}
		}

		// only send if there's any doc keys to query
		if len(docKeys) > 0 {
			err := sendVersionQueryMessage(s.network, clientId, docKeys)
			if err != nil {
				log.Errorf("Synchronizer.handleMessage: Failed to send version query message to client %s: %v", clientId, err)
			} else {
				log.Debugf("Synchronizer.handleMessage: Sent version query message to client %s", clientId)
			}
		}

	case *message.VersionQueryRespMessageV1:
		toUpsert := make(map[string][]byte)
		toDelete := make([]string, 0)
		for docKey, vvBytes := range msg.Responses {
			docKeyBytes := util.String2Bytes(docKey)
			collection, err := key_utils.GetCollectionNameFromKey(docKeyBytes)
			if err != nil {
				log.Errorf("Synchronizer.handleMessage: Failed to get collection name from doc key %s: %v", docKey, err)
				continue
			}
			docId, err := key_utils.GetDocIdFromKey(docKeyBytes)
			if err != nil {
				log.Errorf("Synchronizer.handleMessage: Failed to get doc id from doc key %s: %v", docKey, err)
				continue
			}
			if vvBytes == nil || len(vvBytes) == 0 {
				doc, err := s.managedDb.conn.LoadDoc(collection, docId)
				if err != nil {
					log.Errorf("msgHandler: Failed to load doc %s/%s: %v", collection, docId, err)
					continue
				}
				docBytes := doc.ExportSnapshot().Bytes()
				toUpsert[docKey] = docBytes
			} else {
				doc, err := s.managedDb.conn.LoadDoc(collection, docId)
				if err != nil {
					log.Errorf("msgHandler: Failed to load doc %s/%s: %v", collection, docId, err)
					continue
				}
				vv := loro.NewVvFromBytes(loro.NewRustBytesVec(vvBytes))
				updateBytesVec := doc.ExportUpdatesFrom(vv)
				updateBytes := updateBytesVec.Bytes()
				toUpsert[docKey] = updateBytes
			}
		}
		if len(toUpsert) > 0 {
			err := sendPostDocMessage(s.network, clientId, toUpsert, toDelete)
			if err != nil {
				log.Errorf("Synchronizer.handleMessage: Failed to send post doc message to client %s: %v", clientId, err)
			} else {
				log.Debugf("Synchronizer.handleMessage: Sent post doc message to client %s", clientId)
			}
		}

	default:
		log.Warnf("Synchronizer.handleMessage: Received unknown message type: %T, ignore it", msg)
	}
}

func (s *Synchronizer) handleTransactionCommitted(ev *db_conn.TransactionCommittedEvent) {
	// send ack message to transaction committer
	resp := &message.AckTransactionMessageV1{
		TxID: ev.Transaction.TxID,
	}
	respBytes, err := resp.Encode()
	if err != nil {
		log.Errorf("failed to encode transaction ack message: %v", err)
		return
	}
	s.network.Send(ev.Committer, respBytes)
	log.Debugf("Synchronizer.handleTransactionCommitted: Sent transaction ack message to %s", ev.Committer)

	// notify queryManager of the transaction
	// queryManager updates all query results based on the transaction
	// and returns the updates each client should see
	cus := s.managedDb.queryManager.HandleTransaction(ev.Transaction)
	for clientId, cu := range cus {
		// Skip the committer client
		if clientId == ev.Committer {
			continue
		}

		if cu.IsEmpty() {
			log.Debugf("Synchronizer.handleTransactionCommitted: No updates for client %s", clientId)
		} else { // send post doc message to client
			deletedKeys := make([]string, 0, len(cu.Deletes))
			for docKey := range cu.Deletes {
				deletedKeys = append(deletedKeys, docKey)
			}

			syncMsg := &message.PostDocMessageV1{
				Upsert: cu.Updates,
				Delete: deletedKeys,
			}
			syncMsgBytes, err := syncMsg.Encode()
			if err != nil {
				log.Errorf("Synchronizer.handleTransactionCommitted: Failed to encode post doc message: %v, ignore it", err)
				continue
			}

			err = s.network.Send(clientId, syncMsgBytes)
			if err != nil {
				log.Errorf("Synchronizer.handleTransactionCommitted: Failed to send post doc message to %s: %v, ignore it", clientId, err)
			} else {
				log.Debugf("Synchronizer.handleTransactionCommitted: Sent post doc message to %s", clientId)
			}
		}
	}
}

func (s *Synchronizer) handleTransactionRollbacked(ev *db_conn.TransactionRollbackedEvent) {
	// send TransactionFailedMessage to transaction committer
	err := sendTransactionFailedMessage(
		s.network,
		ev.Committer,
		ev.Transaction.TxID,
		ev.Reason,
	)
	if err != nil {
		log.Errorf("failed to send transaction failed message: %v", err)
		return
	}
	log.Debugf("Synchronizer.handleTransactionRollbacked: Sent transaction failed message to %s", ev.Committer)
}

func (s *Synchronizer) handleConnectionClosed(ev network_server.ConnectionClosedEvent) {
	// remove all subscriptions of a client when it disconnects
	log.Debugf("Synchronizer.handleConnectionClosed: Client %s disconnected", ev.ClientId)
	s.managedDb.queryManager.RemoveAllSubscriptedQueries(ev.ClientId)
}

func sendTransactionFailedMessage(network network_server.NetworkProvider, clientId string, txId string, reason error) error {
	resp := &message.TransactionFailedMessageV1{
		TxID:   txId,
		Reason: reason,
	}
	respBytes, err := resp.Encode()
	if err != nil {
		return pe.Errorf("failed to encode transaction failed message: %v", err)
	}
	network.Send(clientId, respBytes)
	return nil
}

func sendVersionQueryMessage(network network_server.NetworkProvider, clientId string, docKeys map[string]struct{}) error {
	vqm := &message.VersionQueryMessageV1{
		Queries: docKeys,
	}
	vqmBytes, err := vqm.Encode()
	if err != nil {
		return pe.Errorf("failed to encode version query message: %v", err)
	}
	network.Send(clientId, vqmBytes)
	return nil
}

func sendPostDocMessage(network network_server.NetworkProvider, clientId string, toUpsert map[string][]byte, toDelete []string) error {
	postDocMsg := &message.PostDocMessageV1{
		Upsert: toUpsert,
		Delete: toDelete,
	}
	postDocMsgBytes, err := postDocMsg.Encode()
	if err != nil {
		return pe.Errorf("failed to encode post doc message: %v", err)
	}
	network.Send(clientId, postDocMsgBytes)
	return nil
}

// ConnectDatabase connects a database
//
// block until the database is running
func (s *Synchronizer) connectDatabase() error {
	subCtx, cancel := context.WithCancel(s.ctx)

	conn, err := s.dbConnector.ConnectWithContext(subCtx, s.dbUrl)
	if err != nil {
		return err
	}
	err = conn.Open()
	if err != nil {
		return err
	}

	// wait for the database to be running
	<-conn.WaitForStatus(db_conn.DbConnStatusRunning)

	queryExecutor := query_executor.NewQueryExecutor(conn)
	permissionProxy, err := permission_proxy.NewPermissionProxy(conn)
	if err != nil {
		return err
	}
	queryManager := NewQueryManager(queryExecutor, permissionProxy)

	// start a goroutine to listen and handle transaction committed / rollbacked events
	go func() {
		committedEb := conn.GetCommittedEb()
		rollbackedEb := conn.GetRollbackedEb()
		for {
			select {
			case <-subCtx.Done():
				return
			case ev := <-committedEb.Subscribe():
				s.handleTransactionCommitted(ev)
			case ev := <-rollbackedEb.Subscribe():
				s.handleTransactionRollbacked(ev)
			}
		}
	}()

	s.managedDb = &ManagedDb{
		conn:            conn,
		ctx:             subCtx,
		cancel:          cancel,
		queryExecutor:   queryExecutor,
		permissionProxy: permissionProxy,
		queryManager:    queryManager,
	}

	return nil
}
