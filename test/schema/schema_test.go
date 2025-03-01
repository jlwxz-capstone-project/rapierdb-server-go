package main

import (
	_ "embed"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/stretchr/testify/assert"
)

//go:embed test_schema1.js
var testSchema1 string

func TestNewDatabaseSchemaFromJSON(t *testing.T) {
	dbSchema, err := storage_engine.NewDatabaseSchemaFromJs(testSchema1)
	assert.NoError(t, err)

	// 验证解析结果
	assert.Equal(t, storage_engine.DATABASE_SCHEMA, storage_engine.GetType(dbSchema))
	assert.Equal(t, "testDB", dbSchema.Name)
	assert.Equal(t, "1.0.0", dbSchema.Version)
	assert.Len(t, dbSchema.Collections, 1)

	// 验证users集合
	usersCollection, ok := dbSchema.Collections["users"]
	assert.True(t, ok)
	assert.Equal(t, storage_engine.COLLECTION_SCHEMA, storage_engine.GetType(usersCollection))
	assert.Equal(t, "users", usersCollection.Name)

	// 验证文档schema的字段
	fields := usersCollection.DocSchema.Fields

	// 验证id字段
	idField, ok := fields["id"].(*storage_engine.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.STRING_SCHEMA, storage_engine.GetType(idField))
	assert.True(t, idField.Unique)
	assert.Equal(t, storage_engine.HASH_INDEX, idField.IndexType)
	assert.False(t, idField.Nullable)

	// 验证name字段
	nameField, ok := fields["name"].(*storage_engine.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.STRING_SCHEMA, storage_engine.GetType(nameField))
	assert.True(t, nameField.Nullable)
	assert.False(t, nameField.Unique)
	assert.Equal(t, storage_engine.NONE_INDEX, nameField.IndexType)

	// 验证age字段
	ageField, ok := fields["age"].(*storage_engine.NumberSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.NUMBER_SCHEMA, storage_engine.GetType(ageField))
	assert.Equal(t, storage_engine.RANGE_INDEX, ageField.IndexType)

	// 验证tags字段
	tagsField, ok := fields["tags"].(*storage_engine.ListSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.LIST_SCHEMA, storage_engine.GetType(tagsField))
	itemSchema, ok := tagsField.ItemSchema.(*storage_engine.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.STRING_SCHEMA, storage_engine.GetType(itemSchema))

	// 验证profile字段
	profileField, ok := fields["profile"].(*storage_engine.ObjectSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.OBJECT_SCHEMA, storage_engine.GetType(profileField))
	addressField, ok := profileField.Shape["address"].(*storage_engine.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.STRING_SCHEMA, storage_engine.GetType(addressField))
	phoneField, ok := profileField.Shape["phone"].(*storage_engine.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.STRING_SCHEMA, storage_engine.GetType(phoneField))
	assert.True(t, phoneField.Nullable)

	// 验证type字段
	typeField, ok := fields["type"].(*storage_engine.EnumSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.ENUM_SCHEMA, storage_engine.GetType(typeField))
	assert.Equal(t, []string{"admin", "user", "guest"}, typeField.Values)

	// 验证description字段
	descField, ok := fields["description"].(*storage_engine.TextSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.TEXT_SCHEMA, storage_engine.GetType(descField))
	assert.Equal(t, storage_engine.FULLTEXT_INDEX, descField.IndexType)

	// 验证其他复杂字段
	preferencesField, ok := fields["preferences"].(*storage_engine.RecordSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.RECORD_SCHEMA, storage_engine.GetType(preferencesField))

	categoriesField, ok := fields["categories"].(*storage_engine.TreeSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.TREE_SCHEMA, storage_engine.GetType(categoriesField))

	sortedItemsField, ok := fields["sortedItems"].(*storage_engine.MovableListSchema)
	assert.True(t, ok)
	assert.Equal(t, storage_engine.MOVABLE_LIST_SCHEMA, storage_engine.GetType(sortedItemsField))
}
