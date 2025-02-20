// Package storage 提供了存储引擎的实现
package storage

import (
	"errors"
	"fmt"
	"sync"

	"github.com/cockroachdb/pebble"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// 存储引擎使用的常量
const (
	// 存储元数据的键
	STORAGE_META_KEY = "ms"
	// 文档键的前缀
	DOC_KEY_PREFIX = "d"
	// 数据库名称的最大字节数
	DB_SIZE_IN_BYTES = 16
	// 集合名称的最大字节数
	COLLECTION_SIZE_IN_BYTES = 16
	// 文档ID的最大字节数
	DOC_ID_SIZE_IN_BYTES = 16
)

// 文档键的总字节数 = 前缀(1) + 数据库名称字节数(16) + 分隔符(1) + 集合名称字节数(16) + 分隔符(1) + 文档ID字节数(16)
var KEY_SIZE_IN_BYTES = 1 + DB_SIZE_IN_BYTES + 1 + COLLECTION_SIZE_IN_BYTES + 1 + DOC_ID_SIZE_IN_BYTES

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
	// 数据库名称超过最大长度时返回
	ErrDbNameTooLarge = errors.New("db name too large")
	// 集合名称超过最大长度时返回
	ErrCollectionNameTooLarge = errors.New("collection name too large")
	// 文档ID超过最大长度时返回
	ErrDocIDTooLarge = errors.New("doc id too large")
)

// 存储引擎事件类型
const (
	// 事务被取消时触发
	STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED = "storage_engine_event_transaction_canceled"
	// 事务提交成功时触发
	STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED = "storage_engine_event_transaction_committed"
	// 事务回滚时触发
	STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED = "storage_engine_event_transaction_rollbacked"
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
type StorageEngine struct {
	mu        sync.RWMutex        // 用于保护并发访问
	pebbleDB  *pebble.DB          // 底层的 PebbleDB 实例
	docsCache *DocsCache          // 文档缓存
	meta      *StorageMeta        // 存储引擎元数据
	eb        *util.EventBus      // 事件总线
	hooks     *StorageEngineHooks // 存储引擎钩子
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

// OpenStorageEngine 打开存储引擎
// path: 存储路径
// options: 存储引擎选项
func OpenStorageEngine(path string, options StorageEngineOptions) (*StorageEngine, error) {
	pebbleDB, err := pebble.Open(path, options.Options)
	if err != nil {
		return nil, err
	}

	storageMeta, err := loadOrCreateStorageMeta(pebbleDB)
	if err != nil {
		return nil, err
	}

	return &StorageEngine{
		mu:        sync.RWMutex{},
		pebbleDB:  pebbleDB,
		docsCache: NewEmptyDocsCache(),
		meta:      storageMeta,
		eb:        util.NewEventBus(),
		hooks:     &StorageEngineHooks{},
	}, nil
}

// loadOrCreateStorageMeta 加载或创建存储元数据
func loadOrCreateStorageMeta(pebbleDB *pebble.DB) (*StorageMeta, error) {
	if pebbleDB == nil {
		return nil, ErrStorageEngineClosed
	}

	// 尝试获取已存在的存储元数据
	storageMetaBytes, _, err := pebbleDB.Get([]byte(STORAGE_META_KEY))
	if err != nil && err != pebble.ErrNotFound {
		return nil, err
	}

	// 如果元数据不存在，创建新的
	if err == pebble.ErrNotFound {
		storageMeta := NewEmptyStorageMeta()
		storageMetaBytes, err := storageMeta.ToBinary()
		if err != nil {
			return nil, err
		}
		pebbleDB.Set([]byte(STORAGE_META_KEY), storageMetaBytes, pebble.Sync)
		return storageMeta, nil
	} else {
		return StorageMetaFromBinary(storageMetaBytes)
	}
}

// LoadDoc 加载指定的文档
// 如果文档在缓存中存在，直接返回缓存的文档
// 否则从存储中加载文档并更新缓存
func (e *StorageEngine) LoadDoc(dbName, collectionName, docID string) (*loro.LoroDoc, error) {
	key, err := CalcDocKey(dbName, collectionName, docID)
	if err != nil {
		return nil, err
	}

	// 先检查缓存
	{
		e.mu.RLock()
		doc, ok := e.docsCache.docs[string(key)]
		e.mu.RUnlock()
		if ok {
			return doc.Doc, nil
		}
	}

	// 缓存未命中，获取写锁
	e.mu.Lock()
	defer e.mu.Unlock()

	// 双重检查，防止在获取锁的过程中其他goroutine已经加载了文档
	if doc, ok := e.docsCache.docs[string(key)]; ok {
		return doc.Doc, nil
	}

	// 从存储加载文档
	snapshot, _, err := e.pebbleDB.Get(key)
	if err != nil {
		return nil, err
	}

	doc := loro.NewLoroDoc()
	doc.Import(snapshot)

	// 更新缓存
	e.docsCache.docs[string(key)] = &LoadedDoc{DocID: docID, Doc: doc}
	return doc, nil
}

// LoadAllDocsInCollection 加载指定集合中的所有文档
// updateCache: 是否更新缓存
func (e *StorageEngine) LoadAllDocsInCollection(dbName, collectionName string, updateCache bool) (map[string]*loro.LoroDoc, error) {
	lowerbound, err := CalcCollectionLowerBound(dbName, collectionName)
	if err != nil {
		return nil, err
	}

	upperbound, err := CalcCollectionUpperBound(dbName, collectionName)
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
func (e *StorageEngine) LoadDocAndFork(dbName, collectionName, docID string) (*loro.LoroDoc, error) {
	doc, err := e.LoadDoc(dbName, collectionName, docID)
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
			e.eb.Publish(STORAGE_ENGINE_EVENT_TRANSACTION_CANCELED, event)
			return fmt.Errorf("%w: %v", ErrTransactionCancelled, err)
		}
	}

	batch := e.pebbleDB.NewBatch()

	// 处理事务中的每个操作
	for _, op := range tr.Operations {
		switch op := op.(type) {
		case InsertOp:
			{
				database := op.Database
				collection := op.Collection
				docID := op.DocID
				key, err := CalcDocKey(database, collection, docID)
				if err != nil {
					return err
				}

				// 检查文档是否已存在
				oldDoc := e.docsCache.Get(key)
				if oldDoc != nil {
					return fmt.Errorf("%w: doc key = %s", ErrInsertExistingDoc, string(key))
				}

				// 更新缓存
				doc := loro.NewLoroDoc()
				doc.Import(op.Snapshot)
				e.docsCache.Set(key, docID, doc)

				// 记录回滚信息
				rb.toDelete = append(rb.toDelete, key)

				// 添加到批处理
				batch.Set(key, op.Snapshot, pebble.Sync)
			}
		case UpdateOp:
			{
				database := op.Database
				collection := op.Collection
				docID := op.DocID
				key, err := CalcDocKey(database, collection, docID)
				if err != nil {
					return err
				}

				// 检查文档是否存在
				doc := e.docsCache.Get(key)
				if doc == nil {
					return fmt.Errorf("%w: doc key = %s", ErrUpdateNonExistentDoc, string(key))
				}

				// 更新缓存
				forkedDoc := doc.Fork()
				doc.Import(op.Update)
				snapshot := doc.ExportSnapshot()

				// 记录回滚信息
				rbAction := [3]any{
					docID,
					key,
					forkedDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// 添加到批处理
				batch.Set(key, snapshot.Bytes(), pebble.Sync)
			}
		case DeleteOp:
			{
				database := op.Database
				collection := op.Collection
				docID := op.DocID
				key, err := CalcDocKey(database, collection, docID)
				if err != nil {
					return err
				}

				// 检查文档是否存在
				oldDoc := e.docsCache.Get(key)
				if oldDoc == nil {
					return fmt.Errorf("%w: doc key = %s", ErrDeleteNonExistentDoc, string(key))
				}

				// 更新缓存
				e.docsCache.Delete(key)

				// 记录回滚信息
				rbAction := [3]any{
					docID,
					key,
					oldDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// 添加到批处理
				batch.Delete(key, pebble.Sync)
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

	if e.pebbleDB == nil {
		return ErrStorageEngineClosed
	}

	if err := EnsureTransactionValid(tr); err != nil {
		return err
	}

	rb := &rollbackInfo{}
	err := e.commitInner(tr, rb)

	if err == nil {
		// 提交成功，发布事件
		event := &TransactionCommittedEvent{
			Committer:   tr.Committer,
			Transaction: tr,
		}
		e.eb.Publish(STORAGE_ENGINE_EVENT_TRANSACTION_COMMITTED, event)
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
		e.eb.Publish(STORAGE_ENGINE_EVENT_TRANSACTION_ROLLBACKED, event)
	}
	return err
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
	err := e.pebbleDB.Close()
	if err != nil {
		return err
	}
	e.pebbleDB = nil
	e.meta = nil
	e.docsCache = nil
	return nil
}

// SetBeforeTransactionHook 设置事务前钩子
func (e *StorageEngine) SetBeforeTransactionHook(hook *func(tr *Transaction) error) {
	e.hooks.BeforeTransaction = hook
}

// Subscribe 订阅指定主题的事件
func (e *StorageEngine) Subscribe(topic string) <-chan any {
	return e.eb.Subscribe(topic)
}

// Unsubscribe 取消订阅指定主题的事件
func (e *StorageEngine) Unsubscribe(topic string, ch <-chan any) {
	e.eb.Unsubscribe(topic, ch)
}
