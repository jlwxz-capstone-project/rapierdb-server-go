package db_conn

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/pebble"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/key_utils"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

const DefaultDocsCacheSize = 1000

type Caches struct {
	docs *LruCache[loro.LoroDoc]
	meta *DatabaseMeta
}

type Locks struct {
	docsCache sync.RWMutex
}

// rollbackInfo stores rollback information
type rollbackInfo struct {
	toDelete []string // Keys to delete
	toUpdate [][2]any // Document info to update, format: [key, doc object]
}

type PebbleDbConnParams struct {
	Path string
}

func (params *PebbleDbConnParams) EnsureDefaults() {}

type PebbleDbConn struct {
	params *PebbleDbConnParams

	pebbleDb *pebble.DB // Underlying PebbleDB instance

	// Cache
	cache Caches

	// Status Related
	status   atomic.Int32
	statusEb *util.EventBus[DbConnStatus]

	// Context
	ctx    context.Context
	cancel context.CancelFunc

	// Locks
	mu Locks

	// Transaction Related Event Bus
	committedEb  *util.EventBus[*TransactionCommittedEvent]
	rollbackedEb *util.EventBus[*TransactionRollbackedEvent]
}

var _ DbConnection = &PebbleDbConn{}

func CreateNewPebbleDb(path string, schema *DatabaseSchema, permissionJs string) error {
	pebbleOpts := pebble.Options{}
	pebbleOpts.EnsureDefaults()
	pebbleOpts.ErrorIfExists = true
	pebbleDb, err := pebble.Open(path, &pebbleOpts)
	if err != nil {
		return err
	}

	dbMeta := &DatabaseMeta{
		databaseSchema: schema,
		permissionJs:   permissionJs,
		createdAt:      uint64(time.Now().Unix()),
	}

	err = writeDatabaseMeta(pebbleDb, dbMeta)
	if err != nil {
		return err
	}

	err = pebbleDb.Close()
	if err != nil {
		return err
	}

	return nil
}

func NewPebbleDbConnWithContext(ctx context.Context, params *PebbleDbConnParams) (*PebbleDbConn, error) {
	subCtx, cancel := context.WithCancel(ctx)

	conn := &PebbleDbConn{
		params:   params,
		pebbleDb: nil, // init later
		cache: Caches{
			docs: NewLruCache[loro.LoroDoc](DefaultDocsCacheSize),
			meta: nil, // init later
		},
		status:   atomic.Int32{},
		statusEb: util.NewEventBus[DbConnStatus](),
		ctx:      subCtx,
		cancel:   cancel,
		mu: Locks{
			docsCache: sync.RWMutex{},
		},
		committedEb:  util.NewEventBus[*TransactionCommittedEvent](),
		rollbackedEb: util.NewEventBus[*TransactionRollbackedEvent](),
	}

	return conn, nil
}

func (conn *PebbleDbConn) Open() (err error) {
	if !conn.swapStatus(DbConnStatusNotReady, DbConnStatusOpening) {
		return pe.Errorf("cannot open pebble db conn: current status = %d", DbConnStatusNotReady)
	}

	defer func() {
		if err != nil {
			conn.setStatus(DbConnStatusClosing)
			if conn.pebbleDb != nil {
				_ = conn.pebbleDb.Close()
			}
			conn.cancel()
			conn.setStatus(DbConnStatusClosed)
		}
	}()

	// open pebble db
	pebbleOpts := pebble.Options{}
	pebbleOpts.EnsureDefaults()
	pebbleOpts.ErrorIfNotExists = true
	pebbleDb, err := pebble.Open(conn.params.Path, &pebbleOpts)
	if err != nil {
		return pe.Wrap(err, "failed to open pebble db")
	}
	conn.pebbleDb = pebbleDb

	// load database meta
	meta, err := loadDatabaseMeta(pebbleDb)
	if err != nil {
		return pe.Wrap(err, "failed to load database meta")
	}
	conn.cache.meta = meta

	// start a goroutine to listen to the context
	// and trigger a close event if the context is cancelled
	go func() {
		<-conn.ctx.Done()
		conn.Close()
	}()

	// swap to DbConnStatusRunning
	if !conn.swapStatus(DbConnStatusOpening, DbConnStatusRunning) {
		return pe.Errorf("cannot open pebble db conn: current status = %d", DbConnStatusNotReady)
	}

	return nil
}

func (conn *PebbleDbConn) Close() error {
	if !conn.swapStatus(DbConnStatusRunning, DbConnStatusClosing) {
		if !conn.swapStatus(DbConnStatusError, DbConnStatusClosing) {
			return pe.Errorf("cannot close pebble db conn: current status = %d", DbConnStatusRunning)
		}
	}

	conn.cancel()
	conn.cache.docs.Clear()
	if conn.pebbleDb != nil {
		_ = conn.pebbleDb.Close()
	}
	conn.setStatus(DbConnStatusClosed)
	return nil
}

func (conn *PebbleDbConn) GetDatabaseMeta() *DatabaseMeta {
	if conn.GetStatus() != DbConnStatusRunning {
		log.Warnf("database is not running, returned meta maybe nil or not up to date")
	}
	return conn.cache.meta
}

func (conn *PebbleDbConn) UpdateSchema(newSchema *DatabaseSchema) error {
	// XXX!
	if conn.GetStatus() != DbConnStatusRunning {
		return pe.Errorf("cannot update schema: current status = %d", conn.GetStatus())
	}
	oldSchema := conn.cache.meta.databaseSchema
	if newSchema.Version <= oldSchema.Version {
		return pe.Errorf("new schema version must be greater than old schema version")
	}
	meta := conn.cache.meta
	meta.databaseSchema = newSchema
	return writeDatabaseMeta(conn.pebbleDb, meta)
}

func (conn *PebbleDbConn) UpdatePermissionJs(newPermissionJs string) error {
	// XXX!
	if conn.GetStatus() != DbConnStatusRunning {
		return pe.Errorf("cannot update permission: current status = %d", conn.GetStatus())
	}
	meta := conn.cache.meta
	meta.permissionJs = newPermissionJs
	return writeDatabaseMeta(conn.pebbleDb, meta)
}

func (conn *PebbleDbConn) LoadDoc(collectionName, docID string) (*loro.LoroDoc, error) {
	status := conn.GetStatus()
	if status != DbConnStatusRunning {
		return nil, pe.Errorf("cannot load doc: current status = %d", status)
	}

	keyBytes, err := key_utils.CalcDocKey(collectionName, docID)
	if err != nil {
		return nil, err
	}

	// Check cache first
	{
		conn.mu.docsCache.RLock()
		doc, ok := conn.cache.docs.Get(string(keyBytes))
		conn.mu.docsCache.RUnlock()
		if ok {
			return doc, nil
		}
	}

	// Cache miss, acquire write lock
	conn.mu.docsCache.Lock()
	defer conn.mu.docsCache.Unlock()

	if doc, ok := conn.cache.docs.Get(string(keyBytes)); ok {
		return doc, nil
	}

	// Load from pebble db
	snapshot, _, err := conn.pebbleDb.Get(keyBytes)
	if err != nil {
		return nil, pe.Errorf("failed to load doc %s from collection %s: %w", docID, collectionName, err)
	}
	doc := loro.NewLoroDoc()
	doc.Import(snapshot)

	conn.cache.docs.Set(string(keyBytes), doc)
	return doc, nil
}

func (conn *PebbleDbConn) LoadCollection(collectionName string) (map[string]*loro.LoroDoc, error) {
	status := conn.GetStatus()
	if status != DbConnStatusRunning {
		return nil, pe.Errorf("cannot load collection: current status = %d", status)
	}

	lowerbound, err := key_utils.CalcCollectionLowerBound(collectionName)
	if err != nil {
		return nil, err
	}

	upperbound, err := key_utils.CalcCollectionUpperBound(collectionName)
	if err != nil {
		return nil, err
	}

	// Create iterator
	iter, err := conn.pebbleDb.NewIter(&pebble.IterOptions{
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
		docId := key_utils.GetDocIdFromKey(key)
		result[docId] = doc

		conn.cache.docs.Set(string(key), doc)
	}
	return result, nil
}

func (conn *PebbleDbConn) InvalidateCache() {
	conn.cache.docs.Clear()
}

func (conn *PebbleDbConn) commitInner(tr *Transaction, rb *rollbackInfo) error {
	batch := conn.pebbleDb.NewBatch()

	for _, op := range tr.Operations {
		switch op := op.(type) {
		case *InsertOp:
			{
				collection := op.Collection
				docID := op.DocID
				keyBytes, err := key_utils.CalcDocKey(collection, docID)
				key := util.Bytes2String(keyBytes)
				if err != nil {
					return err
				}

				// Check if document already exists
				_, ok := conn.cache.docs.Get(key)
				if ok {
					return pe.Errorf("doc already exists: %s", key)
				}

				// Update cache
				doc := loro.NewLoroDoc()
				doc.Import(op.Snapshot)
				conn.cache.docs.Set(key, doc)

				// Record rollback info
				rb.toDelete = append(rb.toDelete, key)

				// Add to batch
				batch.Set(keyBytes, op.Snapshot, pebble.Sync)
			}
		case *UpdateOp:
			{
				collection := op.Collection
				docID := op.DocID
				keyBytes, err := key_utils.CalcDocKey(collection, docID)
				key := util.Bytes2String(keyBytes)
				if err != nil {
					return err
				}

				// Check if document exists
				doc, ok := conn.cache.docs.Get(key)
				if !ok {
					return pe.Errorf("doc does not exist: %s", key)
				}

				// Update cache
				forkedDoc := doc.Fork()
				doc.Import(op.Update)
				snapshot := doc.ExportSnapshot()

				// Record rollback info
				rbAction := [2]any{
					key,
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
				keyBytes, err := key_utils.CalcDocKey(collection, docID)
				key := util.Bytes2String(keyBytes)
				if err != nil {
					return err
				}

				// Check if document exists
				oldDoc, ok := conn.cache.docs.Get(key)
				if !ok {
					return pe.Errorf("doc does not exist: %s", key)
				}

				// Update cache
				conn.cache.docs.Delete(key)

				// Record rollback info
				rbAction := [2]any{
					key,
					oldDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// Add to batch
				batch.Delete(keyBytes, pebble.Sync)
			}
		}
	}

	return batch.Commit(pebble.Sync)
}

func (conn *PebbleDbConn) Commit(tr *Transaction) error {
	status := conn.GetStatus()
	if status != DbConnStatusRunning {
		return pe.Errorf("cannot commit: current status = %d", status)
	}

	conn.mu.docsCache.Lock()
	defer conn.mu.docsCache.Unlock()

	rb := &rollbackInfo{}
	err := conn.commitInner(tr, rb)

	if err == nil {
		// Commit succeeded, publish event
		event := &TransactionCommittedEvent{
			Committer:   tr.Committer,
			Transaction: tr,
		}
		conn.committedEb.Publish(event)
		return nil
	} else {
		// Commit failed, perform rollback
		for _, key := range rb.toDelete {
			conn.cache.docs.Delete(key)
		}
		for _, action := range rb.toUpdate {
			key := action[0].(string)
			doc := action[1].(*loro.LoroDoc)
			conn.cache.docs.Set(key, doc)
		}
		event := &TransactionRollbackedEvent{
			Committer:   tr.Committer,
			Reason:      err,
			Transaction: tr,
		}
		conn.rollbackedEb.Publish(event)
	}
	return err
}

func (conn *PebbleDbConn) GetCommittedEb() *util.EventBus[*TransactionCommittedEvent] {
	return conn.committedEb
}

func (conn *PebbleDbConn) GetRollbackedEb() *util.EventBus[*TransactionRollbackedEvent] {
	return conn.rollbackedEb
}

func (conn *PebbleDbConn) GetStatus() DbConnStatus {
	return DbConnStatus(conn.status.Load())
}

func (conn *PebbleDbConn) SubscribeStatusChange() <-chan DbConnStatus {
	return conn.statusEb.Subscribe()
}

func (conn *PebbleDbConn) UnsubscribeStatusChange(ch <-chan DbConnStatus) {
	conn.statusEb.Unsubscribe(ch)
}

func (conn *PebbleDbConn) WaitForStatus(targetStatus DbConnStatus) <-chan struct{} {
	statusCh := conn.SubscribeStatusChange()
	cleanup := func() {
		conn.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(conn.GetStatus, targetStatus, statusCh, cleanup, 0)
}

func (conn *PebbleDbConn) setStatus(status DbConnStatus) {
	conn.status.Store(int32(status))
	conn.statusEb.Publish(status)
}

func (conn *PebbleDbConn) swapStatus(from, to DbConnStatus) bool {
	if from == to {
		return false
	}
	if !conn.status.CompareAndSwap(int32(from), int32(to)) {
		return false
	}
	conn.statusEb.Publish(to)
	return true
}

func loadDatabaseMeta(pebbleDB *pebble.DB) (*DatabaseMeta, error) {
	if pebbleDB == nil {
		return nil, pe.Errorf("pebble db is nil")
	}

	// Try to get existing storage metadata
	storageMetaBytes, closer, err := pebbleDB.Get([]byte(key_utils.STORAGE_META_KEY))
	defer closer.Close()

	// Metadata does not exist
	if err != nil {
		return nil, err
	}

	return NewDatabaseMetaFromBytes(storageMetaBytes)
}

func writeDatabaseMeta(pebbleDB *pebble.DB, meta *DatabaseMeta) error {
	if pebbleDB == nil {
		return pe.Errorf("pebble db is nil")
	}

	metaBytes, err := meta.ToBytes()
	if err != nil {
		return err
	}
	return pebbleDB.Set([]byte(key_utils.STORAGE_META_KEY), metaBytes, pebble.Sync)
}
