package main

import (
	"math/rand"
	"testing"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage"
	"github.com/stretchr/testify/assert"
)

// 生成随机字符串的辅助函数
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// 随机生成 CollectionSchema
func randomCollectionSchema() storage.CollectionSchema {
	loadOnTypes := []storage.LoadOnType{storage.LOAD_ON_INIT, storage.LOAD_ON_QUERY}
	return storage.CollectionSchema{
		BaseSchema: storage.BaseSchema{Type: storage.COLLECTION_SCHEMA},
		Name:       randomString(10),
		DocSchema:  storage.BaseSchema{Type: storage.DOC_SCHEMA},
		LoadOn:     loadOnTypes[rand.Intn(len(loadOnTypes))],
	}
}

// 随机生成 DatabaseSchema
func randomDatabaseSchema() storage.DatabaseSchema {
	collections := make(map[string]storage.CollectionSchema)
	collectionCount := rand.Intn(5) + 1 // 1-5个collection

	for i := 0; i < collectionCount; i++ {
		key := randomString(8)
		for _, exists := collections[key]; exists; {
			key = randomString(8)
		}
		collections[key] = randomCollectionSchema()
	}

	return storage.DatabaseSchema{
		BaseSchema:  storage.BaseSchema{Type: storage.DATABASE_SCHEMA},
		Name:        randomString(10),
		Version:     "1.0",
		Collections: collections,
	}
}

// 测试 StorageMeta 的序列化和反序列化
func TestStorageMetaSerialization(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	// 创建一个随机的 StorageMeta
	original := storage.NewEmptyStorageMeta()
	dbCount := rand.Intn(3) + 1 // 1-3个数据库

	for i := 0; i < dbCount; i++ {
		dbName := randomString(8)
		schema := randomDatabaseSchema()
		dbMeta := storage.DatabaseMeta{
			Schema:    schema,
			CreatedAt: uint64(time.Now().Unix()),
		}
		original.DatabaseMetas[dbName] = &dbMeta
	}

	// 序列化
	data, err := original.ToBinary()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	// 反序列化
	restored, err := storage.StorageMetaFromBinary(data)
	assert.NoError(t, err)

	// 验证结果
	assert.Equal(t, len(original.DatabaseMetas), len(restored.DatabaseMetas))
	assert.Equal(t, len(original.DatabaseMetas), len(restored.DatabaseMetas))

	for dbName, originalDB := range original.DatabaseMetas {
		restoredDB, exists := restored.DatabaseMetas[dbName]
		assert.True(t, exists)
		assert.Equal(t, originalDB.CreatedAt, restoredDB.CreatedAt)
		assert.Equal(t, originalDB.Schema.Name, restoredDB.Schema.Name)
		assert.Equal(t, originalDB.Schema.Version, restoredDB.Schema.Version)

		// 验证collections
		assert.Equal(t, len(originalDB.Schema.Collections), len(restoredDB.Schema.Collections))
		for colName, originalCol := range originalDB.Schema.Collections {
			restoredCol, exists := restoredDB.Schema.Collections[colName]
			assert.True(t, exists)
			assert.Equal(t, originalCol.Name, restoredCol.Name)
			assert.Equal(t, originalCol.LoadOn, restoredCol.LoadOn)
			assert.Equal(t, originalCol.Type, restoredCol.Type)
			assert.Equal(t, storage.DOC_SCHEMA, restoredCol.DocSchema.Type)
		}
	}
}

// 测试空的 StorageMeta
func TestEmptyStorageMetaSerialization(t *testing.T) {
	original := storage.NewEmptyStorageMeta()

	data, err := original.ToBinary()
	assert.NoError(t, err)
	assert.NotNil(t, data)

	restored, err := storage.StorageMetaFromBinary(data)
	assert.NoError(t, err)

	assert.Equal(t, len(original.DatabaseMetas), len(restored.DatabaseMetas))
	assert.Equal(t, len(original.DatabaseMetas), len(restored.DatabaseMetas))
}
