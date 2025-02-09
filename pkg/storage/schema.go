package storage

type IndexType string

const (
	HASH_INDEX     IndexType = "hash"
	BTREE_INDEX    IndexType = "btree"
	FULLTEXT_INDEX IndexType = "fulltext"
)

type LoadOnType string

const (
	LOAD_ON_INIT  LoadOnType = "init"
	LOAD_ON_QUERY LoadOnType = "query"
)

type SchemaType string

const (
	ANY_SCHEMA          SchemaType = "any"
	BOOLEAN_SCHEMA      SchemaType = "boolean"
	DATE_SCHEMA         SchemaType = "date"
	ENUM_SCHEMA         SchemaType = "enum"
	LIST_SCHEMA         SchemaType = "list"
	MOVABLE_LIST_SCHEMA SchemaType = "movableList"
	NUMBER_SCHEMA       SchemaType = "number"
	OBJECT_SCHEMA       SchemaType = "object"
	RECORD_SCHEMA       SchemaType = "record"
	STRING_SCHEMA       SchemaType = "string"
	TEXT_SCHEMA         SchemaType = "text"
	TREE_SCHEMA         SchemaType = "tree"
	DOC_SCHEMA          SchemaType = "doc"
	COLLECTION_SCHEMA   SchemaType = "collection"
	DATABASE_SCHEMA     SchemaType = "database"
)

type IndexOptions struct {
	Type IndexType
}

type NullableOptions struct {
	Nullable bool
}

// 如何应用于嵌套的schema？
type UniqueOptions struct {
	Unique bool
}

type BaseSchema struct {
	Type SchemaType
}

type AnySchema struct {
	BaseSchema
	NullableOptions
}

type BooleanSchema struct {
	BaseSchema
	IndexOptions
	NullableOptions
}

type DateSchema struct {
	BaseSchema
	IndexOptions
	NullableOptions
}

type EnumSchema struct {
	BaseSchema
	IndexOptions
	NullableOptions
	Values []string
}

type ListSchema struct {
	BaseSchema
	NullableOptions
	ListItemSchema BaseSchema
}

type MovableListSchema struct {
	BaseSchema
	NullableOptions
	ListItemSchema BaseSchema
}

type NumberSchema struct {
	BaseSchema
	IndexOptions
	NullableOptions
}

type ObjectSchema struct {
	BaseSchema
	NullableOptions
	Shape map[string]BaseSchema
}

type RecordSchema struct {
	BaseSchema
	NullableOptions
	ValueSchema BaseSchema
}

type StringSchema struct {
	IndexOptions
	NullableOptions
}

type TextSchema struct {
	IndexOptions
	NullableOptions
}

type TreeSchema struct {
	IndexOptions
	NullableOptions
	TreeNodeSchema BaseSchema
}

type DocSchema struct {
	BaseSchema
	Fields map[string]BaseSchema
}

type CollectionSchema struct {
	BaseSchema
	Name      string
	DocSchema BaseSchema
	LoadOn    LoadOnType
}

type DatabaseSchema struct {
	BaseSchema
	Name        string
	Version     string
	Collections map[string]CollectionSchema
}
