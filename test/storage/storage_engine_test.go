package main

import (
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage"
	"github.com/stretchr/testify/assert"
)

func TestStorageEngineCRUD(t *testing.T) {
	// 创建临时数据库
	engine := setupEngine(t)
	defer cleanupEngine(t, engine)

	t.Run("应该能插入新文档", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		doc.GetText("name").InsertText("Alice", 0)

		tr := storage.Transaction{
			Operations: []storage.TransactionOp{
				{
					Type: storage.OpInsert,
					InsertOp: storage.InsertOp{
						Database:   "testdb",
						Collection: "users",
						DocID:      "user1",
						Snapshot:   doc.ExportSnapshot().Bytes(),
					},
				},
			},
		}

		err := engine.Commit(&tr)
		assert.NoError(t, err)

		// 验证文档存在
		loadedDoc, err := engine.LoadDoc("testdb", "users", "user1")
		assert.NoError(t, err)
		loadedText := loadedDoc.GetText("name")
		assert.Equal(t, "Alice", loadedText.ToString())
	})

	t.Run("插入已有文档应该报错", func(t *testing.T) {
		loadedDoc, err := engine.LoadDocAndFork("testdb", "users", "user1")
		assert.NoError(t, err)
		loadedText := loadedDoc.GetText("name")
		loadedText.InsertText("Hello ", 0)

		snapshot := loadedDoc.ExportSnapshot().Bytes()
		tr := storage.Transaction{
			Operations: []storage.TransactionOp{
				{
					Type: storage.OpInsert,
					InsertOp: storage.InsertOp{
						Database:   "testdb",
						Collection: "users",
						DocID:      "user1",
						Snapshot:   snapshot,
					},
				},
			},
		}
		assert.Error(t, engine.Commit(&tr))
	})

	t.Run("应该能更新文档", func(t *testing.T) {
		loadedDoc, err := engine.LoadDocAndFork("testdb", "users", "user1")
		assert.NoError(t, err)
		vv := loadedDoc.GetOplogVv()
		loadedText := loadedDoc.GetText("name")
		loadedText.InsertText(" and Bob", loadedText.GetLength())
		assert.Equal(t, "Alice and Bob", loadedText.ToString())

		update := loadedDoc.ExportUpdatesFrom(vv).Bytes()
		tr := storage.Transaction{
			Operations: []storage.TransactionOp{
				{
					Type: storage.OpUpdate,
					UpdateOp: storage.UpdateOp{
						Database:   "testdb",
						Collection: "users",
						DocID:      "user1",
						Update:     update,
					},
				},
			},
		}
		assert.NoError(t, engine.Commit(&tr))
	})

	t.Run("更新不存在的文档应该报错", func(t *testing.T) {
		doc := loro.NewLoroDoc()
		update := doc.ExportAllUpdates().Bytes()
		{
			tr := storage.Transaction{
				Operations: []storage.TransactionOp{
					{
						Type: storage.OpUpdate,
						UpdateOp: storage.UpdateOp{
							Database:   "testdb",
							Collection: "users",
							DocID:      "user2",
							Update:     update,
						},
					},
				},
			}
			assert.Error(t, engine.Commit(&tr))
		}
		{
			tr := storage.Transaction{
				Operations: []storage.TransactionOp{
					{
						Type: storage.OpUpdate,
						UpdateOp: storage.UpdateOp{
							Database:   "testdb",
							Collection: "missingCollection",
							DocID:      "user1",
							Update:     update,
						},
					},
				},
			}
			assert.Error(t, engine.Commit(&tr))
		}
		{
			tr := storage.Transaction{
				Operations: []storage.TransactionOp{
					{
						Type: storage.OpUpdate,
						UpdateOp: storage.UpdateOp{
							Database:   "missingDatabase",
							Collection: "users",
							DocID:      "user1",
							Update:     update,
						},
					},
				},
			}
			assert.Error(t, engine.Commit(&tr))
		}
	})

	t.Run("删除文档", func(t *testing.T) {
		{
			doc := loro.NewLoroDoc()
			doc.GetText("test").InsertText("Hello, World!", 0)
			snapshot := doc.ExportSnapshot().Bytes()
			tr := storage.Transaction{
				Operations: []storage.TransactionOp{
					{
						Type: storage.OpInsert,
						InsertOp: storage.InsertOp{
							Database:   "testdb",
							Collection: "users",
							DocID:      "user2",
							Snapshot:   snapshot,
						},
					},
				},
			}
			assert.NoError(t, engine.Commit(&tr))
		}

		{
			_, err := engine.LoadDoc("testdb", "users", "user2")
			assert.NoError(t, err)
		}

		tr := storage.Transaction{
			Operations: []storage.TransactionOp{
				{
					Type: storage.OpDelete,
					DeleteOp: storage.DeleteOp{
						Database:   "testdb",
						Collection: "users",
						DocID:      "user2",
					},
				},
			},
		}
		assert.NoError(t, engine.Commit(&tr))

		// 验证文档不存在
		_, err := engine.LoadDoc("testdb", "users", "user2")
		assert.Error(t, err)
	})
}

// 辅助函数
func setupEngine(t *testing.T) *storage.StorageEngine {
	opts := storage.DefaultStorageEngineOptions()
	engine, err := storage.OpenStorageEngine(t.TempDir(), *opts)
	assert.NoError(t, err)
	return engine
}

func cleanupEngine(t *testing.T, engine *storage.StorageEngine) {
	assert.NoError(t, engine.Close())
}
