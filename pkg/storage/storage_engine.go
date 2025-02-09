package storage

import (
	"bytes"
	"errors"
	"sync"

	"github.com/cockroachdb/pebble"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
)

const (
	STORAGE_META_KEY = "ms"
	DOC_KEY_PREFIX   = "d"
)

var (
	ErrInsertExistingDoc      = errors.New("insert existing doc")
	ErrUpdateNonExistentDoc   = errors.New("update non-existent doc")
	ErrDeleteNonExistentDoc   = errors.New("delete non-existent doc")
	ErrDatabaseMetaNotFound   = errors.New("database meta not found")
	ErrCollectionMetaNotFound = errors.New("collection meta not found")
	ErrStorageEngineClosed    = errors.New("storage engine closed")
)

type StorageEngine struct {
	mu        sync.RWMutex
	pebbleDB  *pebble.DB
	docsCache *DocsCache
	meta      *StorageMeta
}

type rollbackInfo struct {
	toDelete [][]byte
	toUpdate [][3]interface{} // [string, []byte, *loro.LoroDoc]
}

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
	}, nil
}

func loadOrCreateStorageMeta(pebbleDB *pebble.DB) (*StorageMeta, error) {
	if pebbleDB == nil {
		return nil, ErrStorageEngineClosed
	}

	// 获得 storage meta
	storageMetaBytes, _, err := pebbleDB.Get([]byte(STORAGE_META_KEY))
	if err != nil && err != pebble.ErrNotFound {
		return nil, err
	}

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

func calcDocKey(dbName, collectionName, docID string) []byte {
	var buf bytes.Buffer
	buf.WriteString(DOC_KEY_PREFIX)
	buf.WriteString(dbName)
	buf.WriteByte(':')
	buf.WriteString(collectionName)
	buf.WriteByte(':')
	buf.WriteString(docID)
	return buf.Bytes()
}

func (e *StorageEngine) LoadDoc(dbName, collectionName, docID string) (*loro.LoroDoc, error) {
	key := calcDocKey(dbName, collectionName, docID)

	// 检查缓存
	e.mu.RLock()
	if doc, ok := e.docsCache.docs[string(key)]; ok {
		e.mu.RUnlock()
		return doc.Doc, nil
	}
	e.mu.RUnlock()

	// 从存储加载
	snapshot, _, err := e.pebbleDB.Get(key)
	if err != nil {
		return nil, err
	}

	doc := loro.NewLoroDoc()
	doc.Import(snapshot)

	// 更新缓存
	e.mu.Lock()
	e.docsCache.docs[string(key)] = &LoadedDoc{DocID: docID, Doc: doc}
	e.mu.Unlock()
	return doc, nil
}

func (e *StorageEngine) LoadDocAndFork(dbName, collectionName, docID string) (*loro.LoroDoc, error) {
	doc, err := e.LoadDoc(dbName, collectionName, docID)
	if err != nil {
		return nil, err
	}
	return doc.Fork(), nil
}

func (e *StorageEngine) commitInner(tr *Transaction, rb *rollbackInfo) error {
	batch := e.pebbleDB.NewBatch()

	for _, op := range tr.Operations {
		switch op.Type {
		case OpInsert:
			{
				database := op.InsertOp.Database
				collection := op.InsertOp.Collection
				docID := op.InsertOp.DocID
				key := calcDocKey(database, collection, docID)

				// 不允许插入已存在的文档
				oldDoc := e.docsCache.Get(key)
				if oldDoc != nil {
					return ErrInsertExistingDoc
				}

				// 更新 docsCache
				doc := loro.NewLoroDoc()
				doc.Import(op.InsertOp.Snapshot)
				e.docsCache.Set(key, docID, doc)

				// 计算回滚操作添加到 rb
				rb.toDelete = append(rb.toDelete, key)

				// 写入批量操作
				batch.Set(key, op.InsertOp.Snapshot, pebble.Sync)
			}
		case OpUpdate:
			{
				database := op.UpdateOp.Database
				collection := op.UpdateOp.Collection
				docID := op.UpdateOp.DocID
				key := calcDocKey(database, collection, docID)

				// 不允许更新不存在的文档
				doc := e.docsCache.Get(key)
				if doc == nil {
					return ErrUpdateNonExistentDoc
				}

				// 更新 docsCache
				forkedDoc := doc.Fork()
				doc.Import(op.UpdateOp.Update)
				snapshot := doc.ExportSnapshot()

				// 计算回滚操作添加到 rb
				rbAction := [3]interface{}{
					docID,
					key,
					forkedDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// 写入批量操作
				batch.Set(key, snapshot.Bytes(), pebble.Sync)
			}
		case OpDelete:
			{
				database := op.DeleteOp.Database
				collection := op.DeleteOp.Collection
				docID := op.DeleteOp.DocID
				key := calcDocKey(database, collection, docID)

				// 不允许删除不存在的文档
				oldDoc := e.docsCache.Get(key)
				if oldDoc == nil {
					return ErrDeleteNonExistentDoc
				}

				// 更新 docsCache
				e.docsCache.Delete(key)

				// 计算回滚操作添加到 rb
				rbAction := [3]interface{}{
					docID,
					key,
					oldDoc,
				}
				rb.toUpdate = append(rb.toUpdate, rbAction)

				// 写入批量操作
				batch.Delete(key, pebble.Sync)
			}
		}
	}

	// 写入数据库
	return batch.Commit(pebble.Sync)
}

func (e *StorageEngine) Commit(tr *Transaction) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.pebbleDB == nil {
		return ErrStorageEngineClosed
	}

	rb := &rollbackInfo{}
	err := e.commitInner(tr, rb)
	if err != nil {
		// 发生错误，回滚！
		for _, key := range rb.toDelete {
			e.docsCache.Delete(key)
		}
		for _, action := range rb.toUpdate {
			docID := action[0].(string)
			key := action[1].([]byte)
			doc := action[2].(*loro.LoroDoc)
			e.docsCache.Set(key, docID, doc)
		}
		return err
	}
	return err
}

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
