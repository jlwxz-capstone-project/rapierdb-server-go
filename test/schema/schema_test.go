package main

import (
	_ "embed"
	"testing"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/schema"
	"github.com/stretchr/testify/assert"
)

//go:embed test_schema1.js
var testSchema1 string

func TestNewDatabaseSchemaFromJSON(t *testing.T) {
	dbSchema, err := schema.NewDatabaseSchemaFromJs(testSchema1)
	assert.NoError(t, err)

	// 验证解析结果
	assert.Equal(t, schema.DATABASE_SCHEMA, dbSchema.Type)
	assert.Equal(t, "testDB", dbSchema.Name)
	assert.Equal(t, "1.0.0", dbSchema.Version)
	assert.Len(t, dbSchema.Collections, 1)

	// 验证users集合
	usersCollection, ok := dbSchema.Collections["users"].(*schema.CollectionSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.COLLECTION_SCHEMA, usersCollection.Type)
	assert.Equal(t, "users", usersCollection.Name)

	// 验证文档schema的字段
	fields := usersCollection.DocSchema

	// 验证id字段
	idField, ok := fields["id"].(*schema.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.STRING_SCHEMA, idField.Type)
	assert.True(t, idField.Unique)
	assert.Equal(t, schema.HASH_INDEX, idField.IndexType)
	assert.False(t, idField.Nullable)

	// 验证name字段
	nameField, ok := fields["name"].(*schema.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.STRING_SCHEMA, nameField.Type)
	assert.True(t, nameField.Nullable)
	assert.False(t, nameField.Unique)
	assert.Equal(t, schema.NONE_INDEX, nameField.IndexType)

	// 验证age字段
	ageField, ok := fields["age"].(*schema.NumberSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.NUMBER_SCHEMA, ageField.Type)
	assert.Equal(t, schema.RANGE_INDEX, ageField.IndexType)

	// 验证tags字段
	tagsField, ok := fields["tags"].(*schema.ListSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.LIST_SCHEMA, tagsField.Type)
	itemSchema, ok := tagsField.ItemSchema.(*schema.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.STRING_SCHEMA, itemSchema.Type)

	// 验证profile字段
	profileField, ok := fields["profile"].(*schema.ObjectSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.OBJECT_SCHEMA, profileField.Type)
	addressField, ok := profileField.Shape["address"].(*schema.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.STRING_SCHEMA, addressField.Type)
	phoneField, ok := profileField.Shape["phone"].(*schema.StringSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.STRING_SCHEMA, phoneField.Type)
	assert.True(t, phoneField.Nullable)

	// 验证type字段
	typeField, ok := fields["type"].(*schema.EnumSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.ENUM_SCHEMA, typeField.Type)
	assert.Equal(t, []string{"admin", "user", "guest"}, typeField.Values)

	// 验证description字段
	descField, ok := fields["description"].(*schema.TextSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.TEXT_SCHEMA, descField.Type)
	assert.Equal(t, schema.FULLTEXT_INDEX, descField.IndexType)

	// 验证其他复杂字段
	preferencesField, ok := fields["preferences"].(*schema.RecordSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.RECORD_SCHEMA, preferencesField.Type)

	categoriesField, ok := fields["categories"].(*schema.TreeSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.TREE_SCHEMA, categoriesField.Type)

	sortedItemsField, ok := fields["sortedItems"].(*schema.MovableListSchema)
	assert.True(t, ok)
	assert.Equal(t, schema.MOVABLE_LIST_SCHEMA, sortedItemsField.Type)
}
