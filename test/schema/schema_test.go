package main

import (
	_ "embed"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/stretchr/testify/assert"
)

//go:embed test_schema1.js
var testSchema1 string

func TestNewDatabaseSchemaFromJSON(t *testing.T) {
	dbSchema, err := db_conn.NewDatabaseSchemaFromJs(testSchema1)
	assert.NoError(t, err)

	// 验证解析结果
	assert.Equal(t, db_conn.DATABASE_SCHEMA, db_conn.GetType(dbSchema))
	assert.Equal(t, "testDB", dbSchema.Name)
	assert.Equal(t, "1.0.0", dbSchema.Version)
	assert.Len(t, dbSchema.Collections, 1)

	// 验证users集合
	usersCollection, ok := dbSchema.Collections["users"]
	assert.True(t, ok)
	assert.Equal(t, db_conn.COLLECTION_SCHEMA, db_conn.GetType(usersCollection))
	assert.Equal(t, "users", usersCollection.Name)

	// 验证文档schema的字段
	fields := usersCollection.DocSchema.Fields

	// 验证id字段
	idField, ok := fields["id"].(*db_conn.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.STRING_SCHEMA, db_conn.GetType(idField))
	assert.True(t, idField.Unique)
	assert.Equal(t, db_conn.HASH_INDEX, idField.IndexType)
	assert.False(t, idField.Nullable)

	// 验证name字段
	nameField, ok := fields["name"].(*db_conn.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.STRING_SCHEMA, db_conn.GetType(nameField))
	assert.True(t, nameField.Nullable)
	assert.False(t, nameField.Unique)
	assert.Equal(t, db_conn.NONE_INDEX, nameField.IndexType)

	// 验证age字段
	ageField, ok := fields["age"].(*db_conn.NumberSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.NUMBER_SCHEMA, db_conn.GetType(ageField))
	assert.Equal(t, db_conn.RANGE_INDEX, ageField.IndexType)

	// 验证tags字段
	tagsField, ok := fields["tags"].(*db_conn.ListSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.LIST_SCHEMA, db_conn.GetType(tagsField))
	itemSchema, ok := tagsField.ItemSchema.(*db_conn.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.STRING_SCHEMA, db_conn.GetType(itemSchema))

	// 验证profile字段
	profileField, ok := fields["profile"].(*db_conn.ObjectSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.OBJECT_SCHEMA, db_conn.GetType(profileField))
	addressField, ok := profileField.Shape["address"].(*db_conn.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.STRING_SCHEMA, db_conn.GetType(addressField))
	phoneField, ok := profileField.Shape["phone"].(*db_conn.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.STRING_SCHEMA, db_conn.GetType(phoneField))
	assert.True(t, phoneField.Nullable)

	// 验证type字段
	typeField, ok := fields["type"].(*db_conn.EnumSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.ENUM_SCHEMA, db_conn.GetType(typeField))
	assert.Equal(t, []string{"admin", "user", "guest"}, typeField.Values)

	// 验证description字段
	descField, ok := fields["description"].(*db_conn.TextSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.TEXT_SCHEMA, db_conn.GetType(descField))
	assert.Equal(t, db_conn.FULLTEXT_INDEX, descField.IndexType)

	// 验证其他复杂字段
	preferencesField, ok := fields["preferences"].(*db_conn.RecordSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.RECORD_SCHEMA, db_conn.GetType(preferencesField))

	categoriesField, ok := fields["categories"].(*db_conn.TreeSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.TREE_SCHEMA, db_conn.GetType(categoriesField))

	sortedItemsField, ok := fields["sortedItems"].(*db_conn.MovableListSchema)
	assert.True(t, ok)
	assert.Equal(t, db_conn.MOVABLE_LIST_SCHEMA, db_conn.GetType(sortedItemsField))
}
