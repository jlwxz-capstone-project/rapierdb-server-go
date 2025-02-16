package schema

import (
	_ "embed"
	"errors"
	"fmt"
	"strings"

	"github.com/dop251/goja"
)

type IndexType string

const (
	NONE_INDEX     IndexType = "none"
	HASH_INDEX     IndexType = "hash"
	RANGE_INDEX    IndexType = "range"
	FULLTEXT_INDEX IndexType = "fulltext"
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

type AnySchema struct {
	Type     SchemaType `json:"type"`
	Nullable bool       `json:"nullable"`
}

type BooleanSchema struct {
	Type      SchemaType `json:"type"`
	Nullable  bool       `json:"nullable"`
	Unique    bool       `json:"unique"`
	IndexType IndexType  `json:"indexType"`
}

type DateSchema struct {
	Type      SchemaType `json:"type"`
	Nullable  bool       `json:"nullable"`
	Unique    bool       `json:"unique"`
	IndexType IndexType  `json:"indexType"`
}

type EnumSchema struct {
	Type      SchemaType `json:"type"`
	Values    []string   `json:"values"`
	Nullable  bool       `json:"nullable"`
	Unique    bool       `json:"unique"`
	IndexType IndexType  `json:"indexType"`
}

type ListSchema struct {
	Type       SchemaType  `json:"type"`
	Nullable   bool        `json:"nullable"`
	ItemSchema interface{} `json:"itemSchema"`
}

type MovableListSchema struct {
	Type       SchemaType  `json:"type"`
	Nullable   bool        `json:"nullable"`
	ItemSchema interface{} `json:"itemSchema"`
}

type NumberSchema struct {
	Type      SchemaType `json:"type"`
	Nullable  bool       `json:"nullable"`
	Unique    bool       `json:"unique"`
	IndexType IndexType  `json:"indexType"`
}

type ObjectSchema struct {
	Type     SchemaType             `json:"type"`
	Nullable bool                   `json:"nullable"`
	Shape    map[string]interface{} `json:"shape"`
}

type RecordSchema struct {
	Type        SchemaType  `json:"type"`
	Nullable    bool        `json:"nullable"`
	ValueSchema interface{} `json:"valueSchema"`
}

type StringSchema struct {
	Type      SchemaType `json:"type"`
	Nullable  bool       `json:"nullable"`
	Unique    bool       `json:"unique"`
	IndexType IndexType  `json:"indexType"`
}

type TextSchema struct {
	Type      SchemaType `json:"type"`
	Nullable  bool       `json:"nullable"`
	IndexType IndexType  `json:"indexType"`
}

type TreeSchema struct {
	Type           SchemaType  `json:"type"`
	Nullable       bool        `json:"nullable"`
	TreeNodeSchema interface{} `json:"treeNodeSchema"`
}

type DocSchema struct {
	Type   SchemaType             `json:"type"`
	Fields map[string]interface{} `json:"fields"`
}

type CollectionSchema struct {
	Type      SchemaType             `json:"type"`
	Name      string                 `json:"name"`
	DocSchema map[string]interface{} `json:"docSchema"`
}

type DatabaseSchema struct {
	Type        SchemaType             `json:"type"`
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Collections map[string]interface{} `json:"collections"`
}

//go:embed schema_builder.js
var schemaBuilderScript string

var ErrInvalidDatabaseSchema = errors.New("invalid database schema")

func NewDatabaseSchemaFromJs(js string) (*DatabaseSchema, error) {
	vm := goja.New()
	js = schemaBuilderScript + "\nvar schema = " + strings.Trim(js, "\n ") + "\nschema = schema.toJSON()"
	vm.RunString(js)
	ret := vm.Get("schema").Export()
	jsonSchema, ok := ret.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid schema", ErrInvalidDatabaseSchema)
	}
	schema, err := NewDatabaseSchemaFromJSON(jsonSchema)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDatabaseSchema, err)
	}
	return schema, nil
}

func NewDatabaseSchemaFromJSON(data map[string]interface{}) (*DatabaseSchema, error) {
	schemaType, ok := data["type"].(string)
	if !ok || SchemaType(schemaType) != DATABASE_SCHEMA {
		return nil, fmt.Errorf("%w: `type` of database schema must be \"%s\"", ErrInvalidDatabaseSchema, DATABASE_SCHEMA)
	}

	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: `name` of database schema is required", ErrInvalidDatabaseSchema)
	}

	version, ok := data["version"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: `version` of database schema is required", ErrInvalidDatabaseSchema)
	}

	collectionsData, ok := data["collections"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: `collections` of database schema is required", ErrInvalidDatabaseSchema)
	}

	collections := make(map[string]interface{})
	for key, val := range collectionsData {
		collectionData, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: `collections` must be a object (go map[string]interface{})", ErrInvalidDatabaseSchema)
		}
		collection, err := parseCollectionSchema(collectionData)
		if err != nil {
			return nil, fmt.Errorf("invalid collection %q: %w", key, err)
		}
		collections[key] = collection
	}

	return &DatabaseSchema{
		Type:        DATABASE_SCHEMA,
		Name:        name,
		Version:     version,
		Collections: collections,
	}, nil
}

func parseCollectionSchema(data map[string]interface{}) (*CollectionSchema, error) {
	schemaType, ok := data["type"].(string)
	if !ok || SchemaType(schemaType) != COLLECTION_SCHEMA {
		return nil, fmt.Errorf("%w: invalid collection schema type", ErrInvalidDatabaseSchema)
	}

	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: collection name is required", ErrInvalidDatabaseSchema)
	}

	docSchemaData, ok := data["docSchema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid doc schema", ErrInvalidDatabaseSchema)
	}

	docSchema, err := parseDocSchema(docSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDatabaseSchema, err)
	}

	return &CollectionSchema{
		Type:      COLLECTION_SCHEMA,
		Name:      name,
		DocSchema: docSchema.Fields,
	}, nil
}

func parseDocSchema(data map[string]interface{}) (*DocSchema, error) {
	schemaType, ok := data["type"].(string)
	if !ok || SchemaType(schemaType) != DOC_SCHEMA {
		return nil, fmt.Errorf("%w: invalid doc schema type", ErrInvalidDatabaseSchema)
	}

	fieldsData, ok := data["fields"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid fields data", ErrInvalidDatabaseSchema)
	}

	fields := make(map[string]interface{})
	for key, val := range fieldsData {
		fieldData, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: invalid field data for key: %s", ErrInvalidDatabaseSchema, key)
		}
		field, err := parseFieldSchema(fieldData)
		if err != nil {
			return nil, fmt.Errorf("%w: error parsing field %s: %v", ErrInvalidDatabaseSchema, key, err)
		}
		fields[key] = field
	}

	return &DocSchema{
		Type:   DOC_SCHEMA,
		Fields: fields,
	}, nil
}

func parseFieldSchema(data map[string]interface{}) (interface{}, error) {
	schemaType, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: field type is required", ErrInvalidDatabaseSchema)
	}

	switch SchemaType(schemaType) {
	case ANY_SCHEMA:
		return parseAnySchema(data)
	case BOOLEAN_SCHEMA:
		return parseBooleanSchema(data)
	case DATE_SCHEMA:
		return parseDateSchema(data)
	case ENUM_SCHEMA:
		return parseEnumSchema(data)
	case LIST_SCHEMA:
		return parseListSchema(data)
	case MOVABLE_LIST_SCHEMA:
		return parseMovableListSchema(data)
	case NUMBER_SCHEMA:
		return parseNumberSchema(data)
	case OBJECT_SCHEMA:
		return parseObjectSchema(data)
	case RECORD_SCHEMA:
		return parseRecordSchema(data)
	case STRING_SCHEMA:
		return parseStringSchema(data)
	case TEXT_SCHEMA:
		return parseTextSchema(data)
	case TREE_SCHEMA:
		return parseTreeSchema(data)
	default:
		return nil, fmt.Errorf("%w: unsupported field schema type: %s", ErrInvalidDatabaseSchema, schemaType)
	}
}

// 解析基础类型的辅助函数
func parseAnySchema(data map[string]interface{}) (*AnySchema, error) {
	nullable, _ := data["nullable"].(bool)
	return &AnySchema{
		Type:     ANY_SCHEMA,
		Nullable: nullable,
	}, nil
}

func parseBooleanSchema(data map[string]interface{}) (*BooleanSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &BooleanSchema{
		Type:      BOOLEAN_SCHEMA,
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseDateSchema(data map[string]interface{}) (*DateSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &DateSchema{
		Type:      DATE_SCHEMA,
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseEnumSchema(data map[string]interface{}) (*EnumSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)

	valuesInterface, ok := data["values"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid enum values", ErrInvalidDatabaseSchema)
	}

	values := make([]string, len(valuesInterface))
	for i, v := range valuesInterface {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("%w: enum value must be string", ErrInvalidDatabaseSchema)
		}
		values[i] = str
	}

	return &EnumSchema{
		Type:      ENUM_SCHEMA,
		Values:    values,
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseListSchema(data map[string]interface{}) (*ListSchema, error) {
	nullable, _ := data["nullable"].(bool)

	itemSchemaData, ok := data["itemSchema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid item schema", ErrInvalidDatabaseSchema)
	}

	itemSchema, err := parseFieldSchema(itemSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing item schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &ListSchema{
		Type:       LIST_SCHEMA,
		Nullable:   nullable,
		ItemSchema: itemSchema,
	}, nil
}

func parseMovableListSchema(data map[string]interface{}) (*MovableListSchema, error) {
	nullable, _ := data["nullable"].(bool)

	itemSchemaData, ok := data["itemSchema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid item schema", ErrInvalidDatabaseSchema)
	}

	itemSchema, err := parseFieldSchema(itemSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing item schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &MovableListSchema{
		Type:       MOVABLE_LIST_SCHEMA,
		Nullable:   nullable,
		ItemSchema: itemSchema,
	}, nil
}

func parseNumberSchema(data map[string]interface{}) (*NumberSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &NumberSchema{
		Type:      NUMBER_SCHEMA,
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseObjectSchema(data map[string]interface{}) (*ObjectSchema, error) {
	nullable, _ := data["nullable"].(bool)

	shapeData, ok := data["shape"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid shape data", ErrInvalidDatabaseSchema)
	}

	shape := make(map[string]interface{})
	for key, val := range shapeData {
		fieldData, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("%w: invalid field data for key: %s", ErrInvalidDatabaseSchema, key)
		}
		field, err := parseFieldSchema(fieldData)
		if err != nil {
			return nil, fmt.Errorf("%w: error parsing field %s: %v", ErrInvalidDatabaseSchema, key, err)
		}
		shape[key] = field
	}

	return &ObjectSchema{
		Type:     OBJECT_SCHEMA,
		Nullable: nullable,
		Shape:    shape,
	}, nil
}

func parseRecordSchema(data map[string]interface{}) (*RecordSchema, error) {
	nullable, _ := data["nullable"].(bool)

	valueSchemaData, ok := data["valueSchema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid value schema", ErrInvalidDatabaseSchema)
	}

	valueSchema, err := parseFieldSchema(valueSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing value schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &RecordSchema{
		Type:        RECORD_SCHEMA,
		Nullable:    nullable,
		ValueSchema: valueSchema,
	}, nil
}

func parseStringSchema(data map[string]interface{}) (*StringSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &StringSchema{
		Type:      STRING_SCHEMA,
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseTextSchema(data map[string]interface{}) (*TextSchema, error) {
	nullable, _ := data["nullable"].(bool)
	indexType, _ := data["indexType"].(string)
	return &TextSchema{
		Type:      TEXT_SCHEMA,
		Nullable:  nullable,
		IndexType: IndexType(indexType),
	}, nil
}

func parseTreeSchema(data map[string]interface{}) (*TreeSchema, error) {
	nullable, _ := data["nullable"].(bool)

	treeNodeSchemaData, ok := data["treeNodeSchema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: invalid tree node schema", ErrInvalidDatabaseSchema)
	}

	treeNodeSchema, err := parseFieldSchema(treeNodeSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing tree node schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &TreeSchema{
		Type:           TREE_SCHEMA,
		Nullable:       nullable,
		TreeNodeSchema: treeNodeSchema,
	}, nil
}
