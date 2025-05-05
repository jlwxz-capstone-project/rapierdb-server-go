package main

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/key_utils"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestKeyUtils(t *testing.T) {
	t.Run("calcDocKey 应该正确计算文档键值", func(t *testing.T) {
		key, err := key_utils.CalcDocKey("users", "doc1")
		assert.NoError(t, err)
		assert.Equal(t, "users", key_utils.GetCollectionNameFromKey(key))
		assert.Equal(t, "doc1", key_utils.GetDocIdFromKey(key))
	})

	t.Run("calcDocKey 应该检查字段长度限制", func(t *testing.T) {
		// 集合名称太长
		_, err := key_utils.CalcDocKey("very_very_very_long_collection_name", "doc1")
		assert.Error(t, err)

		// 文档ID太长
		_, err = key_utils.CalcDocKey("users", "very_very_very_long_document_id")
		assert.Error(t, err)
	})

	t.Run("calcCollectionLowerBound 和 calcCollectionUpperBound 应该正确计算范围", func(t *testing.T) {
		lower, err := key_utils.CalcCollectionLowerBound("users")
		assert.NoError(t, err)

		upper, err := key_utils.CalcCollectionUpperBound("users")
		assert.NoError(t, err)

		// 验证下界和上界包含相同的数据库和集合名称
		assert.Equal(t, "users", key_utils.GetCollectionNameFromKey(lower))
		assert.Equal(t, "users", key_utils.GetCollectionNameFromKey(upper))

		// 验证下界的文档ID部分全为0
		docIdLower := key_utils.GetDocIdFromKey(lower)
		for _, b := range []byte(docIdLower) {
			assert.Equal(t, byte(0), b)
		}

		// 验证上界的文档ID部分全为0xFF
		docIdUpper := key_utils.GetDocIdFromKey(upper)
		for _, b := range []byte(docIdUpper) {
			assert.Equal(t, byte(0xFF), b)
		}
	})
}

func TestStorageEngineCRUD(t *testing.T) {
	// 创建临时数据库
	engine := setupConn(t)
	defer cleanupEngine(t, engine)

	t.Run("应该能插入新文档", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		doc.GetText("name").InsertText("Alice", 0)

		tr := db_conn.Transaction{
			TxID:      "11111111-1111-1111-1111-111111111111",
			Committer: "test-client",
			Operations: []db_conn.TransactionOp{
				&db_conn.InsertOp{
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
		loadedDoc, err := engine.LoadDoc("users", "user1")
		assert.NoError(t, err)
		loadedDoc = loadedDoc.Fork()
		loadedText := loadedDoc.GetText("name")
		loadedText.InsertText("Hello ", 0)

		snapshot := loadedDoc.ExportSnapshot().Bytes()
		tr := db_conn.Transaction{
			TxID:      "22222222-2222-2222-2222-222222222222",
			Committer: "test-client",
			Operations: []db_conn.TransactionOp{
				&db_conn.InsertOp{
					Collection: "users",
					DocID:      "user1",
					Snapshot:   snapshot,
				},
			},
		}
		assert.Error(t, engine.Commit(&tr))
	})

	t.Run("应该能更新文档", func(t *testing.T) {
		loadedDoc, err := engine.LoadDoc("users", "user1")
		assert.NoError(t, err)
		loadedDoc = loadedDoc.Fork()
		vv := loadedDoc.GetOplogVv()
		loadedText := loadedDoc.GetText("name")
		loadedText.InsertText(" and Bob", loadedText.GetLength())
		assert.Equal(t, "Alice and Bob", util.Must(loadedText.ToString()))

		update := loadedDoc.ExportUpdatesFrom(vv).Bytes()
		tr := db_conn.Transaction{
			TxID:      "33333333-3333-3333-3333-333333333333",
			Committer: "test-client",
			Operations: []db_conn.TransactionOp{
				&db_conn.UpdateOp{
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
		tr := db_conn.Transaction{
			TxID:      "44444444-4444-4444-4444-444444444444",
			Committer: "test-client",
			Operations: []db_conn.TransactionOp{
				&db_conn.UpdateOp{
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
			tr := db_conn.Transaction{
				TxID:      "77777777-7777-7777-7777-777777777777",
				Committer: "test-client",
				Operations: []db_conn.TransactionOp{
					&db_conn.InsertOp{
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

		tr := db_conn.Transaction{
			TxID:      "88888888-8888-8888-8888-888888888888",
			Committer: "test-client",
			Operations: []db_conn.TransactionOp{
				&db_conn.DeleteOp{
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
			docs, err := engine.LoadCollection("empty_collection")
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

			tr := db_conn.Transaction{
				TxID:      "99999999-9999-9999-9999-99999999999" + user.id[len(user.id)-1:],
				Committer: "test-client",
				Operations: []db_conn.TransactionOp{
					&db_conn.InsertOp{
						Collection: "test_users",
						DocID:      user.id,
						Snapshot:   doc.ExportSnapshot().Bytes(),
					},
				},
			}
			assert.NoError(t, engine.Commit(&tr))
		}

		// 测试加载所有文档
		docs, err := engine.LoadCollection("test_users")
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
	engine := setupConn(t)
	defer cleanupEngine(t, engine)

	t.Run("事务事件应该被正确触发", func(t *testing.T) {
		// 订阅事件
		committedCh := engine.GetCommittedEb().Subscribe()
		rollbackedCh := engine.GetRollbackedEb().Subscribe()
		defer engine.GetCommittedEb().Unsubscribe(committedCh)
		defer engine.GetRollbackedEb().Unsubscribe(rollbackedCh)

		// 测试事务提交事件
		{
			doc := loro.NewLoroDoc()
			doc.GetText("name").InsertText("Success", 0)
			tr := db_conn.Transaction{
				TxID:      "cccccccc-cccc-cccc-cccc-cccccccccccc",
				Committer: "test-client",
				Operations: []db_conn.TransactionOp{
					&db_conn.InsertOp{
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
				op, ok := event.Transaction.Operations[0].(*db_conn.InsertOp)
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
			tr := db_conn.Transaction{
				TxID:      "dddddddd-dddd-dddd-dddd-dddddddddddd",
				Committer: "test-client",
				Operations: []db_conn.TransactionOp{
					&db_conn.InsertOp{
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
				op, ok := event.Transaction.Operations[0].(*db_conn.InsertOp)
				assert.True(t, ok)
				assert.Equal(t, "success_doc", op.DocID)
			case <-time.After(time.Second):
				t.Fatal("未收到事务回滚事件")
			}
		}
	})
}

// 辅助函数
func setupConn(t *testing.T) db_conn.DbConnection {
	dbPath := t.TempDir()
	fmt.Println("dbPath", dbPath)
	dbSchema := db_conn.DatabaseSchema{
		Name:        "testdb",
		Version:     "1.0.0",
		Collections: map[string]*db_conn.CollectionSchema{},
	}
	dbPermissionsJs := `
	Permission.create({
		version: "1.0.0",
		rules: {},
	});
	`
	err := db_conn.CreateNewPebbleDb(dbPath, &dbSchema, dbPermissionsJs)
	assert.NoError(t, err)
	opts := db_conn.PebbleDbConnParams{
		Path: dbPath,
	}
	opts.EnsureDefaults()
	conn, err := db_conn.NewPebbleDbConnWithContext(context.Background(), &opts)
	assert.NoError(t, err)
	err = conn.Open()
	assert.NoError(t, err)
	return conn
}

func cleanupEngine(t *testing.T, conn db_conn.DbConnection) {
	assert.NoError(t, conn.Close())
}
