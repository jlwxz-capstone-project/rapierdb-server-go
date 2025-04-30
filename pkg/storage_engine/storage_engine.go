package storage_engine

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	pe "github.com/pkg/errors"

	"sync/atomic"

	"github.com/cockroachdb/pebble"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// Constants used by the storage engine
const (
	// Key for storing metadata
	STORAGE_META_KEY = "m"
	// Prefix for document keys
	DOC_KEY_PREFIX = "d"
	// Maximum bytes for collection name
	COLLECTION_SIZE_IN_BYTES = 16
	// Maximum bytes for document ID
	DOC_ID_SIZE_IN_BYTES = 16
)

// Total bytes for a document key = prefix(1) + db name bytes(16) + sep(1) + collection name bytes(16) + sep(1) + doc id bytes(16)
var KEY_SIZE_IN_BYTES = 1 + COLLECTION_SIZE_IN_BYTES + 1 + DOC_ID_SIZE_IN_BYTES

// Storage engine event topics
const (
	STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED  = "transaction_committed"
	STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED   = "transaction_canceled"
	STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED = "transaction_rollbacked"
	STORAGE_ENGINE_EVENT_STATUS_CHANGED         = "storage_engine_status_changed"
)

// Storage engine status constants
type StorageEngineStatus int32

const (
	StorageEngineStatusNotStarted StorageEngineStatus = 0
	StorageEngineStatusOpening    StorageEngineStatus = 1
	StorageEngineStatusOpen       StorageEngineStatus = 2
	StorageEngineStatusClosing    StorageEngineStatus = 3
	StorageEngineStatusClosed     StorageEngineStatus = 4
)

func (s StorageEngineStatus) String() string {
	switch s {
	case StorageEngineStatusClosed:
		return "closed"
	case StorageEngineStatusOpening:
		return "opening"
	case StorageEngineStatusOpen:
		return "open"
	case StorageEngineStatusClosing:
		return "closing"
	default:
		return "unknown"
	}
}

type TransactionCanceledEvent struct {
	Committer   string
	Reason      error
	Transaction *Transaction
}

type TransactionCommittedEvent struct {
	Committer   string
	Transaction *Transaction
}

type TransactionRollbackedEvent struct {
	Committer   string
	Reason      error
	Transaction *Transaction
}

// StorageEngine is the main struct for the storage engine.
// Each storage engine manages only one RapierDB database.
// To manage multiple databases, create multiple storage engines.
type StorageEngine struct {
	pebbleDB  *pebble.DB            // Underlying PebbleDB instance
	docsCache *DocsCache            // Document cache
	meta      *DatabaseMeta         // Storage engine metadata
	hooks     *StorageEngineHooks   // Storage engine hooks
	opts      *StorageEngineOptions // Storage engine options

	// Event buses for each event type
	committedEb  *util.EventBus[*TransactionCommittedEvent]
	canceledEb   *util.EventBus[*TransactionCanceledEvent]
	rollbackedEb *util.EventBus[*TransactionRollbackedEvent]

	// Locks
	mu struct {
		docsCache sync.RWMutex
	}

	// Status-related fields
	status   atomic.Int32
	statusEb *util.EventBus[StorageEngineStatus]

	ctx    context.Context
	cancel context.CancelFunc
}

// StorageEngineHooks defines hook functions for the storage engine
type StorageEngineHooks struct {
	// BeforeTransaction is called before a transaction starts.
	// If it returns an error, the transaction will be cancelled.
	BeforeTransaction *func(tr *Transaction) error
}

// rollbackInfo stores rollback information
type rollbackInfo struct {
	toDelete [][]byte // Keys to delete
	toUpdate [][3]any // Document info to update, format: [docID, key, doc object]
}

// CreateNewDatabase creates a new database
// path: storage path for the database, must not already exist
func CreateNewDatabase(path string, schema *DatabaseSchema, permissions string) error {
	pebbleOpts := &pebble.Options{}
	pebbleOpts.EnsureDefaults()
	pebbleOpts.ErrorIfExists = true
	pebbleDB, err := pebble.Open(path, pebbleOpts)
	if err != nil {
		return err
	}

	databaseMeta := &DatabaseMeta{
		DatabaseSchema: schema,
		Permissions:    permissions,
		CreatedAt:      uint64(time.Now().Unix()),
	}

	err = writeDatabaseMeta(pebbleDB, databaseMeta)
	if err != nil {
		return err
	}

	err = pebbleDB.Close()
	if err != nil {
		return err
	}

	return nil
}

// OpenStorageEngine opens a storage engine
// path: storage path
// options: storage engine options
func OpenStorageEngine(opts *StorageEngineOptions) (*StorageEngine, error) {
	return OpenStorageEngineWithContext(opts, context.Background())
}

// OpenStorageEngineWithContext opens a storage engine with the specified context
// It ensures proper status transitions and resource cleanup even if initialization fails.
func OpenStorageEngineWithContext(opts *StorageEngineOptions, ctx context.Context) (engine *StorageEngine, err error) {
	subCtx, cancel := context.WithCancel(ctx)

	engine = &StorageEngine{
		pebbleDB:  nil, // Will be initialized below
		docsCache: nil, // Will be initialized below
		meta:      nil, // Will be initialized below
		hooks:     &StorageEngineHooks{},
		opts:      opts,

		committedEb:  util.NewEventBus[*TransactionCommittedEvent](),
		canceledEb:   util.NewEventBus[*TransactionCanceledEvent](),
		rollbackedEb: util.NewEventBus[*TransactionRollbackedEvent](),

		status:   atomic.Int32{},
		statusEb: util.NewEventBus[StorageEngineStatus](),

		ctx:    subCtx,
		cancel: cancel, // Assign cancel function
	}

	// If err != nil, do cleanup in defer
	defer func() {
		if err != nil { // An error occurred during initialization.
			if engine.pebbleDB != nil {
				_ = engine.pebbleDB.Close()
			}
			engine.setStatus(StorageEngineStatusClosed)
			engine.cancel()
			engine = nil
		}
	}()

	// Indicate that the engine is attempting to open.
	engine.setStatus(StorageEngineStatusOpening)

	// Open the underlying Pebble database.
	var pebbleDB *pebble.DB
	pebbleOpts := opts.GetPebbleOpts()
	pebbleDB, err = pebble.Open(opts.Path, pebbleOpts)
	if err != nil {
		return nil, pe.Wrap(err, "failed to open pebble db")
	}
	// Assign pebbleDB to the engine only after successful opening.
	engine.pebbleDB = pebbleDB

	// Load the database metadata from the Pebble database.
	var meta *DatabaseMeta
	meta, err = loadDatabaseMeta(pebbleDB)
	if err != nil {
		return nil, pe.Wrap(err, "failed to load database metadata")
	}

	// Assign meta to the engine only after successful loading.
	engine.meta = meta

	// Initialize the document cache
	engine.docsCache = NewEmptyDocsCache()

	// Open a goroutine to listen to the context
	// And trigger a close event if the context is cancelled
	go func() {
		<-subCtx.Done()
		engine.Close()
	}()

	// If all initialization steps succeeded, set the status to Open.
	engine.setStatus(StorageEngineStatusOpen)

	return engine, nil
}

// loadDatabaseMeta loads database metadata
func loadDatabaseMeta(pebbleDB *pebble.DB) (*DatabaseMeta, error) {
	if pebbleDB == nil {
		return nil, pe.Errorf("pebble db is nil")
	}

	// Try to get existing storage metadata
	storageMetaBytes, closer, err := pebbleDB.Get([]byte(STORAGE_META_KEY))
	defer closer.Close()

	// Metadata does not exist
	if errors.Is(err, pebble.ErrNotFound) {
		return nil, pe.Errorf("database meta not found")
	} else if err != nil {
		return nil, err
	} else {
		return NewDatabaseMetaFromBytes(storageMetaBytes)
	}
}

// writeDatabaseMeta writes database metadata to the Pebble database
func writeDatabaseMeta(pebbleDB *pebble.DB, meta *DatabaseMeta) error {
	if pebbleDB == nil {
		return pe.Errorf("pebble db is nil")
	}

	metaBytes, err := meta.ToBytes()
	if err != nil {
		return err
	}
	return pebbleDB.Set([]byte(STORAGE_META_KEY), metaBytes, pebble.Sync)
}

// LoadDoc loads the specified document.
// If the document exists in the cache, returns the cached document.
// Otherwise, loads the document from storage and updates the cache.
func (e *StorageEngine) LoadDoc(collectionName, docID string) (*loro.LoroDoc, error) {
	if e.GetStatus() != StorageEngineStatusOpen {
		return nil, pe.Errorf("storage engine is not open")
	}

	keyBytes, err := CalcDocKey(collectionName, docID)
	if err != nil {
		return nil, err
	}

	// Check cache first
	{
		e.mu.docsCache.RLock()
		doc, ok := e.docsCache.Get(keyBytes)
		e.mu.docsCache.RUnlock()
		if ok {
			return doc, nil
		}
	}

	// Cache miss, acquire write lock
	e.mu.docsCache.Lock()
	defer e.mu.docsCache.Unlock()

	// Double check in case another goroutine loaded the doc while acquiring the lock
	if doc, ok := e.docsCache.Get(keyBytes); ok {
		return doc, nil
	}

	// Load document from storage
	snapshot, _, err := e.pebbleDB.Get(keyBytes)
	if err != nil {
		return nil, pe.WithStack(fmt.Errorf("failed to load doc %s from collection %s: %w", docID, collectionName, err))
	}

	doc := loro.NewLoroDoc()
	doc.Import(snapshot)

	// Update cache
	e.docsCache.Set(keyBytes, docID, doc)
	return doc, nil
}

// LoadAllDocsInCollection loads all documents in the specified collection.
// updateCache: whether to update the cache
// Returns: map from document ID to document
func (e *StorageEngine) LoadAllDocsInCollection(collectionName string, updateCache bool) (map[string]*loro.LoroDoc, error) {
	if e.GetStatus() != StorageEngineStatusOpen {
		return nil, pe.Errorf("storage engine is not open")
	}

	lowerbound, err := CalcCollectionLowerBound(collectionName)
	if err != nil {
		return nil, err
	}

	upperbound, err := CalcCollectionUpperBound(collectionName)
	if err != nil {
		return nil, err
	}

	// Create iterator
	iter, err := e.pebbleDB.NewIter(&pebble.IterOptions{
		LowerBound: lowerbound,
		UpperBound: upperbound,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// Iterate all documents in the collection
	result := make(map[string]*loro.LoroDoc)
	for iter.First(); iter.Valid(); iter.Next() {
		key := iter.Key()
		snapshot := iter.Value()
		doc := loro.NewLoroDoc()
		doc.Import(snapshot)
		docId := GetDocIdFromKey(key)
		result[docId] = doc

		if updateCache {
			e.docsCache.Set(key, docId, doc)
		}
	}
	return result, nil
}

// LoadDocAndFork loads a document and creates a fork
func (e *StorageEngine) LoadDocAndFork(collectionName, docID string) (*loro.LoroDoc, error) {
	if e.GetStatus() != StorageEngineStatusOpen {
		return nil, pe.Errorf("storage engine is not open")
	}

	doc, err := e.LoadDoc(collectionName, docID)
	if err != nil {
		return nil, err
	}
	return doc.Fork(), nil
}

// commitInner performs the internal logic for committing a transaction
func (e *StorageEngine) commitInner(tr *Transaction, rb *rollbackInfo) error {
	// Call before-transaction hook
	if e.hooks.BeforeTransaction != nil {
		err := (*e.hooks.BeforeTransaction)(tr)
		if err != nil {
			event := &TransactionCanceledEvent{
				Committer:   tr.Committer,
				Reason:      err,
				Transaction: tr,
			}
			e.canceledEb.Publish(event)
			return pe.Errorf("transaction cancelled: %v", err)
		}
	}

	batch := e.pebbleDB.NewBatch()

	// Process each operation in the transaction
	for _, op := range tr.Operations {
		switch op := op.(type) {
		case *InsertOp:
			{
				collection := op.Collection
				docID := op.DocID
				keyBytes, err := CalcDocKey(collection, docID)
				key := util.Bytes2String(keyBytes)
				if err != nil {
					return err
				}

				// Check if document already exists
				_, ok := e.docsCache.Get(keyBytes)
				if ok {
					return pe.Errorf("doc already exists: %s", key)
				}

				// Update cache
				doc := loro.NewLoroDoc()
				doc.Import(op.Snapshot)
				e.docsCache.Set(keyBytes, docID, doc)

				// Record rollback info
				rb.toDelete = append(rb.toDelete, keyBytes)

				// Add to batch
				batch.Set(keyBytes, op.Snapshot, pebble.Sync)
			}
		case *UpdateOp:
			{
				collection := op.Collection
				docID := op.DocID
				keyBytes, err := CalcDocKey(collection, docID)
				key := util.Bytes2String(keyBytes)
				if err != nil {
					return err
				}

				// Check if document exists
				doc, ok := e.docsCache.Get(keyBytes)
				if !ok {
					return pe.Errorf("doc does not exist: %s", key)
				}

				// Update cache
				forkedDoc := doc.Fork()
				doc.Import(op.Update)
				snapshot := doc.ExportSnapshot()

				// Record rollback info
				rbAction := [3]any{
					docID,
					keyBytes,
					forkedDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// Add to batch
				batch.Set(keyBytes, snapshot.Bytes(), pebble.Sync)
			}
		case *DeleteOp:
			{
				collection := op.Collection
				docID := op.DocID
				keyBytes, err := CalcDocKey(collection, docID)
				key := util.Bytes2String(keyBytes)
				if err != nil {
					return err
				}

				// Check if document exists
				oldDoc, ok := e.docsCache.Get(keyBytes)
				if !ok {
					return pe.Errorf("doc does not exist: %s", key)
				}

				// Update cache
				e.docsCache.Delete(keyBytes)

				// Record rollback info
				rbAction := [3]any{
					docID,
					keyBytes,
					oldDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// Add to batch
				batch.Delete(keyBytes, pebble.Sync)
			}
		}
	}

	// Commit the batch
	return batch.Commit(pebble.Sync)
}

// Commit commits a transaction
//
//   - If the database is closed, returns ErrStorageEngineClosed
//   - If the transaction is committed successfully, both the cache and underlying storage are updated, and a commit event is triggered
//   - If the transaction commit fails, both the cache and storage are rolled back, a rollback event is triggered,
//     and the error that caused the failure is returned
//   - Before committing, the beforeTransactionHook is called.
//     You can perform permission checks in this hook. If it returns an error, the transaction is cancelled.
func (e *StorageEngine) Commit(tr *Transaction) error {
	if e.GetStatus() != StorageEngineStatusOpen {
		return pe.Errorf("storage engine is not open")
	}

	e.mu.docsCache.Lock()
	defer e.mu.docsCache.Unlock()

	// Check if transaction is valid
	if err := EnsureTransactionValid(tr); err != nil {
		return err
	}

	// Ensure the transaction's target is the current database
	if tr.TargetDatabase != e.meta.DatabaseSchema.Name {
		return fmt.Errorf("%w: target database = %s, expected = %s", ErrTransactionInvalid, tr.TargetDatabase, e.meta.DatabaseSchema.Name)
	}

	rb := &rollbackInfo{}
	err := e.commitInner(tr, rb)

	if err == nil {
		// Commit succeeded, publish event
		event := &TransactionCommittedEvent{
			Committer:   tr.Committer,
			Transaction: tr,
		}
		e.committedEb.Publish(event)
		return nil
	} else {
		// Commit failed, perform rollback
		for _, key := range rb.toDelete {
			e.docsCache.Delete(key)
		}
		for _, action := range rb.toUpdate {
			docID := action[0].(string)
			key := action[1].([]byte)
			doc := action[2].(*loro.LoroDoc)
			e.docsCache.Set(key, docID, doc)
		}
		event := &TransactionRollbackedEvent{
			Committer:   tr.Committer,
			Reason:      err,
			Transaction: tr,
		}
		e.rollbackedEb.Publish(event)
	}
	return err
}

// GetStatus returns the current status of the storage engine
func (e *StorageEngine) GetStatus() StorageEngineStatus {
	return StorageEngineStatus(e.status.Load())
}

// setStatus sets the storage engine status and notifies subscribers
func (e *StorageEngine) setStatus(status StorageEngineStatus) {
	oldStatus := e.status.Load()
	e.status.Store(int32(status))

	if oldStatus != int32(status) {
		e.statusEb.Publish(status)
	}
}

// SubscribeStatusChange subscribes to status change events
func (e *StorageEngine) SubscribeStatusChange() <-chan StorageEngineStatus {
	return e.statusEb.Subscribe()
}

// UnsubscribeStatusChange unsubscribes from status change events
func (e *StorageEngine) UnsubscribeStatusChange(ch <-chan StorageEngineStatus) {
	e.statusEb.Unsubscribe(ch)
}

// WaitForStatus waits for the storage engine to reach the specified status.
// Returns a channel that will be notified when the target status is reached.
func (e *StorageEngine) WaitForStatus(targetStatus StorageEngineStatus) <-chan struct{} {
	statusCh := e.SubscribeStatusChange()
	cleanup := func() {
		e.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(e.GetStatus, targetStatus, statusCh, cleanup, 0)
}

// Close closes the storage engine
func (e *StorageEngine) Close() error {
	status := e.GetStatus()
	if status == StorageEngineStatusOpen {
		e.setStatus(StorageEngineStatusClosing)

		// Cancel context
		e.cancel()

		// Close pebble db
		_ = e.pebbleDB.Close()
		e.pebbleDB = nil
		e.meta = nil
		e.docsCache = nil

		e.setStatus(StorageEngineStatusClosed)
	}
	return nil
}

func (e *StorageEngine) IsValidCollection(collectionName string) bool {
	_, ok := e.meta.DatabaseSchema.Collections[collectionName]
	return ok
}

// SetBeforeTransactionHook sets the before-transaction hook
func (e *StorageEngine) SetBeforeTransactionHook(hook *func(tr *Transaction) error) {
	e.hooks.BeforeTransaction = hook
}

// SubscribeCommitted subscribes to transaction committed events
func (e *StorageEngine) SubscribeCommitted() <-chan *TransactionCommittedEvent {
	return e.committedEb.Subscribe()
}

// UnsubscribeCommitted unsubscribes from transaction committed events
func (e *StorageEngine) UnsubscribeCommitted(ch <-chan *TransactionCommittedEvent) {
	e.committedEb.Unsubscribe(ch)
}

// SubscribeCanceled subscribes to transaction canceled events
func (e *StorageEngine) SubscribeCanceled() <-chan *TransactionCanceledEvent {
	return e.canceledEb.Subscribe()
}

// UnsubscribeCanceled unsubscribes from transaction canceled events
func (e *StorageEngine) UnsubscribeCanceled(ch <-chan *TransactionCanceledEvent) {
	e.canceledEb.Unsubscribe(ch)
}

// SubscribeRollbacked subscribes to transaction rollbacked events
func (e *StorageEngine) SubscribeRollbacked() <-chan *TransactionRollbackedEvent {
	return e.rollbackedEb.Subscribe()
}

// UnsubscribeRollbacked unsubscribes from transaction rollbacked events
func (e *StorageEngine) UnsubscribeRollbacked(ch <-chan *TransactionRollbackedEvent) {
	e.rollbackedEb.Unsubscribe(ch)
}

func (e *StorageEngine) GetDbPath() string {
	return e.opts.Path
}

func (e *StorageEngine) GetPermissionsJs() string {
	return e.meta.Permissions
}
