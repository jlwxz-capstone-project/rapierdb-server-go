package storage_engine

import (
	"errors"
	"fmt"
	"sync"
	"time"

	pe "github.com/pkg/errors"

	"github.com/cockroachdb/pebble"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// 存储引擎使用的常量
const (
	// 存储元数据的键
	STORAGE_META_KEY = "m"
	// 文档键的前缀
	DOC_KEY_PREFIX = "d"
	// 集合名称的最大字节数
	COLLECTION_SIZE_IN_BYTES = 16
	// 文档ID的最大字节数
	DOC_ID_SIZE_IN_BYTES = 16
)

// 文档键的总字节数 = 前缀(1) + 数据库名称字节数(16) + 分隔符(1) + 集合名称字节数(16) + 分隔符(1) + 文档ID字节数(16)
var KEY_SIZE_IN_BYTES = 1 + COLLECTION_SIZE_IN_BYTES + 1 + DOC_ID_SIZE_IN_BYTES

// 存储引擎可能返回的错误
var (
	// 尝试插入已存在的文档时返回
	ErrInsertExistingDoc = errors.New("insert existing doc")
	// 尝试更新不存在的文档时返回
	ErrUpdateNonExistentDoc = errors.New("update non-existent doc")
	// 尝试删除不存在的文档时返回
	ErrDeleteNonExistentDoc = errors.New("delete non-existent doc")
	// 数据库元数据不存在时返回
	ErrDatabaseMetaNotFound = errors.New("database meta not found")
	// 集合元数据不存在时返回
	ErrCollectionMetaNotFound = errors.New("collection meta not found")
	// 存储引擎已关闭时返回
	ErrStorageEngineClosed = errors.New("storage engine closed")
	// 事务被取消时返回
	ErrTransactionCancelled = errors.New("transaction cancelled")
	// 集合名称超过最大长度时返回
	ErrCollectionNameTooLarge = errors.New("collection name too large")
	// 文档ID超过最大长度时返回
	ErrDocIDTooLarge = errors.New("doc id too large")
)

// 存储引擎事件主题
const (
	STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED  = "transaction_committed"
	STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED   = "transaction_canceled"
	STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED = "transaction_rollbacked"
	STORAGE_ENGINE_EVENT_STATUS_CHANGED         = "storage_engine_status_changed"
)

// 存储引擎状态常量
type StorageEngineStatus string

const (
	StorageEngineStatusClosed  StorageEngineStatus = "closed"
	StorageEngineStatusOpening StorageEngineStatus = "opening"
	StorageEngineStatusOpen    StorageEngineStatus = "open"
	StorageEngineStatusClosing StorageEngineStatus = "closing"
)

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

// StorageEngine 是存储引擎的主要结构体
// 每个存储引擎只能管理一个 RapierDB 数据库
// 如果需要管理多个数据库，需要创建多个存储引擎
type StorageEngine struct {
	mu        sync.RWMutex          // 用于保护并发访问
	pebbleDB  *pebble.DB            // 底层的 PebbleDB 实例
	docsCache *DocsCache            // 文档缓存
	meta      *DatabaseMeta         // 存储引擎元数据
	hooks     *StorageEngineHooks   // 存储引擎钩子
	opts      *StorageEngineOptions // 存储引擎选项

	// 每类事件对应的消息总线
	committedEb  *util.EventBus[*TransactionCommittedEvent]
	canceledEb   *util.EventBus[*TransactionCanceledEvent]
	rollbackedEb *util.EventBus[*TransactionRollbackedEvent]

	// 状态相关字段
	status     StorageEngineStatus
	statusLock sync.RWMutex
	statusEb   *util.EventBus[StorageEngineStatus]
}

// StorageEngineHooks 定义了存储引擎的钩子函数
type StorageEngineHooks struct {
	// BeforeTransaction 在事务开始前调用
	// 如果返回错误，事务将被取消
	BeforeTransaction *func(tr *Transaction) error
}

// rollbackInfo 存储回滚信息
type rollbackInfo struct {
	toDelete [][]byte // 需要删除的键
	toUpdate [][3]any // 需要更新的文档信息，格式为 [文档ID, 键, 文档对象]
}

// CreateNewDatabase 创建一个新的数据库
// path: 数据库的存储路径，要求不能有已经存在的数据库
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

// OpenStorageEngine 打开存储引擎
// path: 存储路径
// options: 存储引擎选项
func OpenStorageEngine(opts *StorageEngineOptions) (*StorageEngine, error) {
	engine := &StorageEngine{
		mu:        sync.RWMutex{},
		pebbleDB:  nil, // 下面初始化
		docsCache: nil, // 下面初始化
		meta:      nil, // 下面初始化
		hooks:     &StorageEngineHooks{},
		opts:      opts,

		committedEb:  util.NewEventBus[*TransactionCommittedEvent](),
		canceledEb:   util.NewEventBus[*TransactionCanceledEvent](),
		rollbackedEb: util.NewEventBus[*TransactionRollbackedEvent](),

		status:     StorageEngineStatusClosed,
		statusLock: sync.RWMutex{},
		statusEb:   util.NewEventBus[StorageEngineStatus](),
	}

	engine.setStatus(StorageEngineStatusOpening)

	pebbleOpts := opts.GetPebbleOpts()
	pebbleDB, err := pebble.Open(opts.Path, pebbleOpts)
	if err != nil {
		engine.setStatus(StorageEngineStatusClosed)
		return nil, err
	}
	engine.pebbleDB = pebbleDB

	// 加载数据库元数据
	meta, err := loadDatabaseMeta(pebbleDB)
	if err != nil {
		pebbleDB.Close()
		engine.setStatus(StorageEngineStatusClosed)
		return nil, err
	}
	engine.meta = meta

	// 初始化文档缓存
	engine.docsCache = NewEmptyDocsCache()

	engine.setStatus(StorageEngineStatusOpen)
	return engine, nil
}

// loadDatabaseMeta 加载数据库元数据
func loadDatabaseMeta(pebbleDB *pebble.DB) (*DatabaseMeta, error) {
	if pebbleDB == nil {
		return nil, ErrStorageEngineClosed
	}

	// 尝试获取已存在的存储元数据
	storageMetaBytes, closer, err := pebbleDB.Get([]byte(STORAGE_META_KEY))

	// 元数据不存在
	if err == pebble.ErrNotFound {
		return nil, ErrDatabaseMetaNotFound
	} else if err != nil {
		return nil, err
	} else {
		closer.Close()
		return NewDatabaseMetaFromBytes(storageMetaBytes)
	}
}

func writeDatabaseMeta(pebbleDB *pebble.DB, meta *DatabaseMeta) error {
	metaBytes, err := meta.ToBytes()
	if err != nil {
		return err
	}
	return pebbleDB.Set([]byte(STORAGE_META_KEY), metaBytes, pebble.Sync)
}

// LoadDoc 加载指定的文档
// 如果文档在缓存中存在，直接返回缓存的文档
// 否则从存储中加载文档并更新缓存
func (e *StorageEngine) LoadDoc(collectionName, docID string) (*loro.LoroDoc, error) {
	keyBytes, err := CalcDocKey(collectionName, docID)
	if err != nil {
		return nil, err
	}

	// 先检查缓存
	{
		e.mu.RLock()
		doc, ok := e.docsCache.Get(keyBytes)
		e.mu.RUnlock()
		if ok {
			return doc, nil
		}
	}

	// 缓存未命中，获取写锁
	e.mu.Lock()
	defer e.mu.Unlock()

	// 双重检查，防止在获取锁的过程中其他goroutine已经加载了文档
	if doc, ok := e.docsCache.Get(keyBytes); ok {
		return doc, nil
	}

	// 从存储加载文档
	snapshot, _, err := e.pebbleDB.Get(keyBytes)
	if err != nil {
		return nil, pe.WithStack(fmt.Errorf("failed to load doc %s from collection %s: %w", docID, collectionName, err))
	}

	doc := loro.NewLoroDoc()
	doc.Import(snapshot)

	// 更新缓存
	e.docsCache.Set(keyBytes, docID, doc)
	return doc, nil
}

// LoadAllDocsInCollection 加载指定集合中的所有文档
// updateCache: 是否更新缓存
// 返回值: 文档ID到文档的映射
func (e *StorageEngine) LoadAllDocsInCollection(collectionName string, updateCache bool) (map[string]*loro.LoroDoc, error) {
	lowerbound, err := CalcCollectionLowerBound(collectionName)
	if err != nil {
		return nil, err
	}

	upperbound, err := CalcCollectionUpperBound(collectionName)
	if err != nil {
		return nil, err
	}

	// 创建迭代器
	iter, err := e.pebbleDB.NewIter(&pebble.IterOptions{
		LowerBound: lowerbound,
		UpperBound: upperbound,
	})
	if err != nil {
		return nil, err
	}
	defer iter.Close()

	// 遍历集合中的所有文档
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

// LoadDocAndFork 加载文档并创建一个分支
func (e *StorageEngine) LoadDocAndFork(collectionName, docID string) (*loro.LoroDoc, error) {
	doc, err := e.LoadDoc(collectionName, docID)
	if err != nil {
		return nil, err
	}
	return doc.Fork(), nil
}

// commitInner 执行事务提交的内部逻辑
func (e *StorageEngine) commitInner(tr *Transaction, rb *rollbackInfo) error {
	// 执行事务前钩子
	if e.hooks.BeforeTransaction != nil {
		err := (*e.hooks.BeforeTransaction)(tr)
		if err != nil {
			event := &TransactionCanceledEvent{
				Committer:   tr.Committer,
				Reason:      err,
				Transaction: tr,
			}
			e.canceledEb.Publish(event)
			return fmt.Errorf("%w: %v", ErrTransactionCancelled, err)
		}
	}

	batch := e.pebbleDB.NewBatch()

	// 处理事务中的每个操作
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

				// 检查文档是否已存在
				_, ok := e.docsCache.Get(keyBytes)
				if ok {
					return fmt.Errorf("%w: doc key = %s", ErrInsertExistingDoc, key)
				}

				// 更新缓存
				doc := loro.NewLoroDoc()
				doc.Import(op.Snapshot)
				e.docsCache.Set(keyBytes, docID, doc)

				// 记录回滚信息
				rb.toDelete = append(rb.toDelete, keyBytes)

				// 添加到批处理
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

				// 检查文档是否存在
				doc, ok := e.docsCache.Get(keyBytes)
				if !ok {
					return fmt.Errorf("%w: doc key = %s", ErrUpdateNonExistentDoc, key)
				}

				// 更新缓存
				forkedDoc := doc.Fork()
				doc.Import(op.Update)
				snapshot := doc.ExportSnapshot()

				// 记录回滚信息
				rbAction := [3]any{
					docID,
					keyBytes,
					forkedDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// 添加到批处理
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

				// 检查文档是否存在
				oldDoc, ok := e.docsCache.Get(keyBytes)
				if !ok {
					return fmt.Errorf("%w: doc key = %s", ErrDeleteNonExistentDoc, key)
				}

				// 更新缓存
				e.docsCache.Delete(keyBytes)

				// 记录回滚信息
				rbAction := [3]any{
					docID,
					keyBytes,
					oldDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// 添加到批处理
				batch.Delete(keyBytes, pebble.Sync)
			}
		}
	}

	// 提交批处理
	return batch.Commit(pebble.Sync)
}

// Commit 提交事务
//
//   - 如果数据库已关闭，则返回 ErrStorageEngineClosed
//   - 如果事务提交成功，缓存和底层存储都会被更新，并触发事务提交事件
//   - 如果事务提交失败，则缓存和底层存储都会被回滚，触发事务回滚事件，
//     并返回导致事务提交失败的错误
//   - 事务提交前，会调用 beforeTransactionHook 钩子函数。
//     可以在这个钩子中执行权限检查等操作。如果返回错误，则事务会被取消
func (e *StorageEngine) Commit(tr *Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	// 存储引擎关闭后不能提交事务
	if e.pebbleDB == nil {
		return ErrStorageEngineClosed
	}

	// 检查事务是否有效
	if err := EnsureTransactionValid(tr); err != nil {
		return err
	}

	// 确认这个事务的目标是当前数据库
	if tr.TargetDatabase != e.meta.DatabaseSchema.Name {
		return fmt.Errorf("%w: target database = %s, expected = %s", ErrTransactionInvalid, tr.TargetDatabase, e.meta.DatabaseSchema.Name)
	}

	rb := &rollbackInfo{}
	err := e.commitInner(tr, rb)

	if err == nil {
		// 提交成功，发布事件
		event := &TransactionCommittedEvent{
			Committer:   tr.Committer,
			Transaction: tr,
		}
		e.committedEb.Publish(event)
		return nil
	} else if !errors.Is(err, ErrTransactionCancelled) {
		// 提交失败，执行回滚
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

// GetStatus 获取存储引擎当前状态
func (e *StorageEngine) GetStatus() StorageEngineStatus {
	e.statusLock.RLock()
	defer e.statusLock.RUnlock()
	return e.status
}

// setStatus 设置存储引擎状态并通知订阅者
func (e *StorageEngine) setStatus(status StorageEngineStatus) {
	e.statusLock.Lock()
	oldStatus := e.status
	e.status = status
	e.statusLock.Unlock()

	// 只有状态发生变化时才发布事件
	if oldStatus != status {
		// 通过事件总线发布状态变更事件
		e.statusEb.Publish(status)
	}
}

// SubscribeStatusChange 订阅状态变更事件
func (e *StorageEngine) SubscribeStatusChange() <-chan StorageEngineStatus {
	return e.statusEb.Subscribe()
}

// UnsubscribeStatusChange 取消订阅状态变更事件
func (e *StorageEngine) UnsubscribeStatusChange(ch <-chan StorageEngineStatus) {
	e.statusEb.Unsubscribe(ch)
}

// WaitForStatus 等待存储引擎达到指定状态，返回一个通道，当达到目标状态时会收到通知
func (e *StorageEngine) WaitForStatus(targetStatus StorageEngineStatus) <-chan struct{} {
	statusCh := e.SubscribeStatusChange()
	cleanup := func() {
		e.UnsubscribeStatusChange(statusCh)
	}
	return util.WaitForStatus(e.GetStatus, targetStatus, statusCh, cleanup, 0)
}

// IsClosed 检查存储引擎是否已关闭
func (e *StorageEngine) IsClosed() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.pebbleDB == nil
}

// Close 关闭存储引擎
func (e *StorageEngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.setStatus(StorageEngineStatusClosing)

	err := e.pebbleDB.Close()
	if err != nil {
		return err
	}
	e.pebbleDB = nil
	e.meta = nil
	e.docsCache = nil

	e.setStatus(StorageEngineStatusClosed)
	return nil
}

func (e *StorageEngine) IsValidCollection(collectionName string) bool {
	_, ok := e.meta.DatabaseSchema.Collections[collectionName]
	return ok
}

// SetBeforeTransactionHook 设置事务前钩子
func (e *StorageEngine) SetBeforeTransactionHook(hook *func(tr *Transaction) error) {
	e.hooks.BeforeTransaction = hook
}

// SubscribeCommitted 订阅事务提交事件
func (e *StorageEngine) SubscribeCommitted() <-chan *TransactionCommittedEvent {
	return e.committedEb.Subscribe()
}

// UnsubscribeCommitted 取消订阅事务提交事件
func (e *StorageEngine) UnsubscribeCommitted(ch <-chan *TransactionCommittedEvent) {
	e.committedEb.Unsubscribe(ch)
}

// SubscribeCanceled 订阅事务取消事件
func (e *StorageEngine) SubscribeCanceled() <-chan *TransactionCanceledEvent {
	return e.canceledEb.Subscribe()
}

// UnsubscribeCanceled 取消订阅事务取消事件
func (e *StorageEngine) UnsubscribeCanceled(ch <-chan *TransactionCanceledEvent) {
	e.canceledEb.Unsubscribe(ch)
}

// SubscribeRollbacked 订阅事务回滚事件
func (e *StorageEngine) SubscribeRollbacked() <-chan *TransactionRollbackedEvent {
	return e.rollbackedEb.Subscribe()
}

// UnsubscribeRollbacked 取消订阅事务回滚事件
func (e *StorageEngine) UnsubscribeRollbacked(ch <-chan *TransactionRollbackedEvent) {
	e.rollbackedEb.Unsubscribe(ch)
}

func (e *StorageEngine) GetDbPath() string {
	return e.opts.Path
}

func (e *StorageEngine) GetPermissionsJs() string {
	return e.meta.Permissions
}
