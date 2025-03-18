package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestKeyUtils(t *testing.T) {
	t.Run("calcDocKey 应该正确计算文档键值", func(t *testing.T) {
		key, err := storage_engine.CalcDocKey("users", "doc1")
		assert.NoError(t, err)
		assert.Equal(t, "users", storage_engine.GetCollectionNameFromKey(key))
		assert.Equal(t, "doc1", storage_engine.GetDocIdFromKey(key))
	})

	t.Run("calcDocKey 应该检查字段长度限制", func(t *testing.T) {
		// 集合名称太长
		_, err := storage_engine.CalcDocKey("very_very_very_long_collection_name", "doc1")
		assert.Error(t, err)

		// 文档ID太长
		_, err = storage_engine.CalcDocKey("users", "very_very_very_long_document_id")
		assert.Error(t, err)
	})

	t.Run("calcCollectionLowerBound 和 calcCollectionUpperBound 应该正确计算范围", func(t *testing.T) {
		lower, err := storage_engine.CalcCollectionLowerBound("users")
		assert.NoError(t, err)

		upper, err := storage_engine.CalcCollectionUpperBound("users")
		assert.NoError(t, err)

		// 验证下界和上界包含相同的数据库和集合名称
		assert.Equal(t, "users", storage_engine.GetCollectionNameFromKey(lower))
		assert.Equal(t, "users", storage_engine.GetCollectionNameFromKey(upper))

		// 验证下界的文档ID部分全为0
		docIdLower := storage_engine.GetDocIdFromKey(lower)
		for _, b := range []byte(docIdLower) {
			assert.Equal(t, byte(0), b)
		}

		// 验证上界的文档ID部分全为0xFF
		docIdUpper := storage_engine.GetDocIdFromKey(upper)
		for _, b := range []byte(docIdUpper) {
			assert.Equal(t, byte(0xFF), b)
		}
	})
}

func TestStorageEngineCRUD(t *testing.T) {
	// 创建临时数据库
	engine := setupEngine(t)
	defer cleanupEngine(t, engine)

	t.Run("应该能插入新文档", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		doc.GetText("name").InsertText("Alice", 0)

		tr := storage_engine.Transaction{
			TxID:           "11111111-1111-1111-1111-111111111111",
			TargetDatabase: "testdb",
			Committer:      "test-client",
			Operations: []any{
				storage_engine.InsertOp{
					Collection: "users",
					DocID:      "user1",
					Snapshot:   doc.ExportSnapshot().Bytes(),
				},
			},
		}

		err := engine.Commit(&tr)
		assert.NoError(t, err)

		// 验证文档存在
		loadedDoc, err := engine.LoadDoc("users", "user1")
		assert.NoError(t, err)
		loadedText := loadedDoc.GetText("name")
		assert.Equal(t, "Alice", util.Must(loadedText.ToString()))
	})

	t.Run("插入已有文档应该报错", func(t *testing.T) {
		loadedDoc, err := engine.LoadDocAndFork("users", "user1")
		assert.NoError(t, err)
		loadedText := loadedDoc.GetText("name")
		loadedText.InsertText("Hello ", 0)

		snapshot := loadedDoc.ExportSnapshot().Bytes()
		tr := storage_engine.Transaction{
			TxID:           "22222222-2222-2222-2222-222222222222",
			TargetDatabase: "testdb",
			Committer:      "test-client",
			Operations: []any{
				storage_engine.InsertOp{
					Collection: "users",
					DocID:      "user1",
					Snapshot:   snapshot,
				},
			},
		}
		assert.Error(t, engine.Commit(&tr))
	})

	t.Run("应该能更新文档", func(t *testing.T) {
		loadedDoc, err := engine.LoadDocAndFork("users", "user1")
		assert.NoError(t, err)
		vv := loadedDoc.GetOplogVv()
		loadedText := loadedDoc.GetText("name")
		loadedText.InsertText(" and Bob", loadedText.GetLength())
		assert.Equal(t, "Alice and Bob", util.Must(loadedText.ToString()))

		update := loadedDoc.ExportUpdatesFrom(vv).Bytes()
		tr := storage_engine.Transaction{
			TxID:           "33333333-3333-3333-3333-333333333333",
			TargetDatabase: "testdb",
			Committer:      "test-client",
			Operations: []any{
				storage_engine.UpdateOp{
					Collection: "users",
					DocID:      "user1",
					Update:     update,
				},
			},
		}
		assert.NoError(t, engine.Commit(&tr))
	})

	t.Run("更新不存在的文档应该报错", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		update := doc.ExportAllUpdates().Bytes()
		tr := storage_engine.Transaction{
			TxID:           "44444444-4444-4444-4444-444444444444",
			TargetDatabase: "testdb",
			Committer:      "test-client",
			Operations: []any{
				storage_engine.UpdateOp{
					Collection: "users",
					DocID:      "xxxxxxx",
					Update:     update,
				},
			},
		}
		assert.Error(t, engine.Commit(&tr))

	})

	t.Run("删除文档", func(t *testing.T) {
		{
			doc := loro.NewLoroDoc()
			doc.GetText("test").InsertText("Hello, World!", 0)
			snapshot := doc.ExportSnapshot().Bytes()
			tr := storage_engine.Transaction{
				TxID:           "77777777-7777-7777-7777-777777777777",
				TargetDatabase: "testdb",
				Committer:      "test-client",
				Operations: []any{
					storage_engine.InsertOp{
						Collection: "users",
						DocID:      "user2",
						Snapshot:   snapshot,
					},
				},
			}
			assert.NoError(t, engine.Commit(&tr))
		}

		{
			_, err := engine.LoadDoc("users", "user2")
			assert.NoError(t, err)
		}

		tr := storage_engine.Transaction{
			TxID:           "88888888-8888-8888-8888-888888888888",
			TargetDatabase: "testdb",
			Committer:      "test-client",
			Operations: []any{
				storage_engine.DeleteOp{
					Collection: "users",
					DocID:      "user2",
				},
			},
		}
		assert.NoError(t, engine.Commit(&tr))

		// 验证文档不存在
		_, err := engine.LoadDoc("users", "user2")
		assert.Error(t, err)
	})

	t.Run("应该能加载集合中的所有文档", func(t *testing.T) {
		// 测试空集合
		{
			docs, err := engine.LoadAllDocsInCollection("empty_collection", false)
			assert.NoError(t, err)
			assert.Empty(t, docs)
		}

		// 准备测试数据：插入多个文档
		users := []struct {
			id   string
			name string
			age  int
		}{
			{"user_a", "Alice", 25},
			{"user_b", "Bob", 30},
			{"user_c", "Charlie", 35},
		}

		for _, user := range users {
			doc := loro.NewLoroDoc()
			doc.GetText("name").UpdateText(user.name)
			doc.GetText("age").UpdateText(fmt.Sprintf("%d", user.age))

			tr := storage_engine.Transaction{
				TxID:           "99999999-9999-9999-9999-99999999999" + user.id[len(user.id)-1:],
				TargetDatabase: "testdb",
				Committer:      "test-client",
				Operations: []any{
					storage_engine.InsertOp{
						Collection: "test_users",
						DocID:      user.id,
						Snapshot:   doc.ExportSnapshot().Bytes(),
					},
				},
			}
			assert.NoError(t, engine.Commit(&tr))
		}

		// 测试加载所有文档
		docs, err := engine.LoadAllDocsInCollection("test_users", false)
		assert.NoError(t, err)
		assert.Len(t, docs, len(users))

		// 验证每个文档的内容
		for _, user := range users {
			doc, exists := docs[user.id]
			assert.True(t, exists)
			assert.Equal(t, user.name, util.Must(doc.GetText("name").ToString()))
			assert.Equal(t, fmt.Sprintf("%d", user.age), util.Must(doc.GetText("age").ToString()))
		}
	})
}

func TestStorageEngineHooksAndEvents(t *testing.T) {
	engine := setupEngine(t)
	defer cleanupEngine(t, engine)

	t.Run("BeforeTransaction hook 应该能阻止事务", func(t *testing.T) {
		// 设置 hook 函数
		hookCalled := false
		hook := func(tr *storage_engine.Transaction) error {
			hookCalled = true
			// 检查事务中是否包含特定的文档ID
			if tr.Operations[0].(storage_engine.InsertOp).DocID == "blocked_doc" {
				return fmt.Errorf("document blocked")
			}
			return nil
		}
		engine.SetBeforeTransactionHook(&hook)

		// 尝试插入被阻止的文档
		doc := loro.NewLoroDoc()
		doc.GetText("name").InsertText("Blocked", 0)
		tr := storage_engine.Transaction{
			TxID:           "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
			TargetDatabase: "testdb",
			Committer:      "test-client",
			Operations: []any{
				storage_engine.InsertOp{
					Collection: "users",
					DocID:      "blocked_doc",
					Snapshot:   doc.ExportSnapshot().Bytes(),
				},
			},
		}

		// 验证事务被阻止
		err := engine.Commit(&tr)
		assert.Error(t, err)
		assert.True(t, hookCalled)
		assert.Contains(t, err.Error(), "document blocked")

		// 验证文档未被插入
		_, err = engine.LoadDoc("users", "blocked_doc")
		assert.Error(t, err)
	})
	t.Run("事务事件应该被正确触发", func(t *testing.T) {
		// 订阅事件
		committedCh := engine.SubscribeCommitted()
		canceledCh := engine.SubscribeCanceled()
		rollbackedCh := engine.SubscribeRollbacked()
		defer engine.UnsubscribeCommitted(committedCh)
		defer engine.UnsubscribeCanceled(canceledCh)
		defer engine.UnsubscribeRollbacked(rollbackedCh)

		// 测试事务取消事件
		{
			// 创建一个新的 loro doc 用于测试
			doc := loro.NewLoroDoc()
			snapshot := doc.ExportSnapshot()

			tr := &storage_engine.Transaction{
				TxID:           "00000000-0000-0000-0000-000000000001",
				TargetDatabase: "testdb",
				Committer:      "test-client",
				Operations: []any{
					storage_engine.InsertOp{
						Collection: "users",
						DocID:      "canceled_doc",
						Snapshot:   snapshot.Bytes(),
					},
				},
			}

			// 设置一个会导致事务取消的 hook
			hook := func(tr *storage_engine.Transaction) error {
				if tr != nil && len(tr.Operations) > 0 {
					if op, ok := tr.Operations[0].(storage_engine.InsertOp); ok && op.DocID == "canceled_doc" {
						return fmt.Errorf("transaction canceled")
					}
				}
				return nil
			}
			engine.SetBeforeTransactionHook(&hook)

			// 提交事务
			err := engine.Commit(tr)
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "transaction canceled")

			// 验证收到取消事件
			select {
			case event := <-canceledCh:
				assert.Equal(t, "test-client", event.Committer)
				assert.Equal(t, "canceled_doc", event.Transaction.Operations[0].(storage_engine.InsertOp).DocID)
			case <-time.After(time.Second):
				t.Fatal("未收到事务取消事件")
			}
		}

		// 测试事务提交事件
		{
			// 移除之前的 hook
			engine.SetBeforeTransactionHook(nil)

			doc := loro.NewLoroDoc()
			doc.GetText("name").InsertText("Success", 0)
			tr := storage_engine.Transaction{
				TxID:           "cccccccc-cccc-cccc-cccc-cccccccccccc",
				TargetDatabase: "testdb",
				Committer:      "test-client",
				Operations: []any{
					storage_engine.InsertOp{
						Collection: "users",
						DocID:      "success_doc",
						Snapshot:   doc.ExportSnapshot().Bytes(),
					},
				},
			}
			engine.Commit(&tr)

			// 验证收到提交事件
			select {
			case event := <-committedCh:
				op, ok := event.Transaction.Operations[0].(storage_engine.InsertOp)
				assert.True(t, ok)
				assert.Equal(t, "success_doc", op.DocID)
			case <-time.After(time.Second):
				t.Fatal("未收到事务提交事件")
			}
		}

		// 测试事务回滚事件（通过尝试插入重复文档触发）
		{
			doc := loro.NewLoroDoc()
			doc.GetText("name").InsertText("Duplicate", 0)
			tr := storage_engine.Transaction{
				TxID:           "dddddddd-dddd-dddd-dddd-dddddddddddd",
				TargetDatabase: "testdb",
				Committer:      "test-client",
				Operations: []any{
					storage_engine.InsertOp{
						Collection: "users",
						DocID:      "success_doc", // 使用已存在的文档ID
						Snapshot:   doc.ExportSnapshot().Bytes(),
					},
				},
			}
			engine.Commit(&tr)

			// 验证收到回滚事件
			select {
			case event := <-rollbackedCh:
				op, ok := event.Transaction.Operations[0].(storage_engine.InsertOp)
				assert.True(t, ok)
				assert.Equal(t, "success_doc", op.DocID)
			case <-time.After(time.Second):
				t.Fatal("未收到事务回滚事件")
			}
		}
	})
}

// 辅助函数
func setupEngine(t *testing.T) *storage_engine.StorageEngine {
	dbPath := t.TempDir()
	fmt.Println("dbPath", dbPath)
	dbSchema := storage_engine.DatabaseSchema{
		Name:        "testdb",
		Version:     "1.0.0",
		Collections: map[string]*storage_engine.CollectionSchema{},
	}
	dbPermissionsJs := `
	Permission.create({
		version: "1.0.0",
		rules: {},
	});
	`
	err := storage_engine.CreateNewDatabase(dbPath, &dbSchema, dbPermissionsJs)
	assert.NoError(t, err)
	opts, err := storage_engine.DefaultStorageEngineOptions(dbPath)
	assert.NoError(t, err)
	engine, err := storage_engine.OpenStorageEngine(opts)
	assert.NoError(t, err)
	return engine
}

func cleanupEngine(t *testing.T, engine *storage_engine.StorageEngine) {
	assert.NoError(t, engine.Close())
}
