package schema

import (
	"fmt"
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

func NewDatabaseSchemaFromJSON(data map[string]interface{}) (*DatabaseSchema, error) {
	schemaType, ok := data["type"].(string)
	if !ok || SchemaType(schemaType) != DATABASE_SCHEMA {
		return nil, fmt.Errorf("invalid database schema type")
	}

	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid database name")
	}

	version, ok := data["version"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid database version")
	}

	collectionsData, ok := data["collections"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid collections data")
	}

	collections := make(map[string]interface{})
	for key, val := range collectionsData {
		collectionData, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid collection data for key: %s", key)
		}
		collection, err := parseCollectionSchema(collectionData)
		if err != nil {
			return nil, fmt.Errorf("error parsing collection %s: %v", key, err)
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
		return nil, fmt.Errorf("invalid collection schema type")
	}

	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid collection name")
	}

	docSchemaData, ok := data["docSchema"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid doc schema data")
	}

	docSchema, err := parseDocSchema(docSchemaData)
	if err != nil {
		return nil, fmt.Errorf("error parsing doc schema: %v", err)
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
		return nil, fmt.Errorf("invalid doc schema type")
	}

	fieldsData, ok := data["fields"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("invalid fields data")
	}

	fields := make(map[string]interface{})
	for key, val := range fieldsData {
		fieldData, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid field data for key: %s", key)
		}
		field, err := parseFieldSchema(fieldData)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s: %v", key, err)
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
		return nil, fmt.Errorf("invalid field schema type")
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
		return nil, fmt.Errorf("unsupported field schema type: %s", schemaType)
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
		return nil, fmt.Errorf("invalid enum values")
	}

	values := make([]string, len(valuesInterface))
	for i, v := range valuesInterface {
		str, ok := v.(string)
		if !ok {
			return nil, fmt.Errorf("enum value must be string")
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
		return nil, fmt.Errorf("invalid item schema")
	}

	itemSchema, err := parseFieldSchema(itemSchemaData)
	if err != nil {
		return nil, fmt.Errorf("error parsing item schema: %v", err)
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
		return nil, fmt.Errorf("invalid item schema")
	}

	itemSchema, err := parseFieldSchema(itemSchemaData)
	if err != nil {
		return nil, fmt.Errorf("error parsing item schema: %v", err)
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
		return nil, fmt.Errorf("invalid shape data")
	}

	shape := make(map[string]interface{})
	for key, val := range shapeData {
		fieldData, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid field data for key: %s", key)
		}
		field, err := parseFieldSchema(fieldData)
		if err != nil {
			return nil, fmt.Errorf("error parsing field %s: %v", key, err)
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
		return nil, fmt.Errorf("invalid value schema")
	}

	valueSchema, err := parseFieldSchema(valueSchemaData)
	if err != nil {
		return nil, fmt.Errorf("error parsing value schema: %v", err)
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
		return nil, fmt.Errorf("invalid tree node schema")
	}

	treeNodeSchema, err := parseFieldSchema(treeNodeSchemaData)
	if err != nil {
		return nil, fmt.Errorf("error parsing tree node schema: %v", err)
	}

	return &TreeSchema{
		Type:           TREE_SCHEMA,
		Nullable:       nullable,
		TreeNodeSchema: treeNodeSchema,
	}, nil
}
