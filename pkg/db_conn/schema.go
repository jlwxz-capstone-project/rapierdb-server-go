package db_conn

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

const (
	ANY_SCHEMA          = "any"
	BOOLEAN_SCHEMA      = "boolean"
	DATE_SCHEMA         = "date"
	ENUM_SCHEMA         = "enum"
	LIST_SCHEMA         = "list"
	MOVABLE_LIST_SCHEMA = "movableList"
	NUMBER_SCHEMA       = "number"
	OBJECT_SCHEMA       = "object"
	RECORD_SCHEMA       = "record"
	STRING_SCHEMA       = "string"
	TEXT_SCHEMA         = "text"
	TREE_SCHEMA         = "tree"
	DOC_SCHEMA          = "doc"
	COLLECTION_SCHEMA   = "collection"
	DATABASE_SCHEMA     = "database"
)

type AnySchema struct {
	Nullable bool `json:"nullable"`
}

type BooleanSchema struct {
	Nullable  bool      `json:"nullable"`
	Unique    bool      `json:"unique"`
	IndexType IndexType `json:"indexType"`
}

type DateSchema struct {
	Nullable  bool      `json:"nullable"`
	Unique    bool      `json:"unique"`
	IndexType IndexType `json:"indexType"`
}

type EnumSchema struct {
	Values    []string  `json:"values"`
	Nullable  bool      `json:"nullable"`
	Unique    bool      `json:"unique"`
	IndexType IndexType `json:"indexType"`
}

type ListSchema struct {
	Nullable   bool `json:"nullable"`
	ItemSchema any  `json:"itemSchema"`
}

type MovableListSchema struct {
	Nullable   bool `json:"nullable"`
	ItemSchema any  `json:"itemSchema"`
}

type NumberSchema struct {
	Nullable  bool      `json:"nullable"`
	Unique    bool      `json:"unique"`
	IndexType IndexType `json:"indexType"`
}

type ObjectSchema struct {
	Nullable bool           `json:"nullable"`
	Shape    map[string]any `json:"shape"`
}

type RecordSchema struct {
	Nullable    bool `json:"nullable"`
	ValueSchema any  `json:"valueSchema"`
}

type StringSchema struct {
	Nullable  bool      `json:"nullable"`
	Unique    bool      `json:"unique"`
	IndexType IndexType `json:"indexType"`
}

type TextSchema struct {
	Nullable  bool      `json:"nullable"`
	IndexType IndexType `json:"indexType"`
}

type TreeSchema struct {
	Nullable       bool `json:"nullable"`
	TreeNodeSchema any  `json:"treeNodeSchema"`
}

type ValueSchema interface {
	AnySchema |
		BooleanSchema |
		DateSchema |
		EnumSchema |
		ListSchema |
		MovableListSchema |
		NumberSchema |
		ObjectSchema |
		RecordSchema |
		StringSchema |
		TextSchema |
		TreeSchema
}

type DocSchema struct {
	Fields map[string]any `json:"fields"`
}

type CollectionSchema struct {
	Name      string     `json:"name"`
	DocSchema *DocSchema `json:"docSchema"`
}

type DatabaseSchema struct {
	Name        string                       `json:"name"`
	Version     string                       `json:"version"`
	Collections map[string]*CollectionSchema `json:"collections"`
}

type Schema interface {
	ValueSchema |
		DocSchema |
		CollectionSchema |
		DatabaseSchema
}

//go:embed schema_builder.js
var schemaBuilderScript string

var ErrInvalidDatabaseSchema = errors.New("invalid database schema")

func GetType[T Schema](s *T) string {
	switch v := any(s).(type) {
	case *AnySchema:
		return ANY_SCHEMA
	case *BooleanSchema:
		return BOOLEAN_SCHEMA
	case *DateSchema:
		return DATE_SCHEMA
	case *EnumSchema:
		return ENUM_SCHEMA
	case *ListSchema:
		return LIST_SCHEMA
	case *MovableListSchema:
		return MOVABLE_LIST_SCHEMA
	case *NumberSchema:
		return NUMBER_SCHEMA
	case *ObjectSchema:
		return OBJECT_SCHEMA
	case *RecordSchema:
		return RECORD_SCHEMA
	case *StringSchema:
		return STRING_SCHEMA
	case *TextSchema:
		return TEXT_SCHEMA
	case *TreeSchema:
		return TREE_SCHEMA
	case *DocSchema:
		return DOC_SCHEMA
	case *CollectionSchema:
		return COLLECTION_SCHEMA
	case *DatabaseSchema:
		return DATABASE_SCHEMA
	default:
		panic(fmt.Sprintf("unknown schema type: %T", v))
	}
}

func NewDatabaseSchemaFromJs(js string) (*DatabaseSchema, error) {
	vm := goja.New()
	js = schemaBuilderScript + "\nvar schema = " + strings.Trim(js, "\n ") + "\nschema = schema.toJSON()"
	vm.RunString(js)
	ret := vm.Get("schema").Export()
	jsonSchema, ok := ret.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid schema", ErrInvalidDatabaseSchema)
	}
	schema, err := NewDatabaseSchemaFromJSON(jsonSchema)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDatabaseSchema, err)
	}
	return schema, nil
}

func NewDatabaseSchemaFromJSON(data map[string]any) (*DatabaseSchema, error) {
	schemaType, ok := data["type"].(string)
	if !ok || schemaType != DATABASE_SCHEMA {
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

	collectionsData, ok := data["collections"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: `collections` of database schema is required", ErrInvalidDatabaseSchema)
	}

	collections := make(map[string]*CollectionSchema)
	for key, val := range collectionsData {
		collectionData, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%w: `collections` must be a object (go map[string]any)", ErrInvalidDatabaseSchema)
		}
		collection, err := parseCollectionSchema(collectionData)
		if err != nil {
			return nil, fmt.Errorf("invalid collection %q: %w", key, err)
		}
		collections[key] = collection
	}

	return &DatabaseSchema{
		Name:        name,
		Version:     version,
		Collections: collections,
	}, nil
}

func parseCollectionSchema(data map[string]any) (*CollectionSchema, error) {
	schemaType, ok := data["type"].(string)
	if !ok || schemaType != COLLECTION_SCHEMA {
		return nil, fmt.Errorf("%w: invalid collection schema type", ErrInvalidDatabaseSchema)
	}

	name, ok := data["name"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: collection name is required", ErrInvalidDatabaseSchema)
	}

	docSchemaData, ok := data["docSchema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid doc schema", ErrInvalidDatabaseSchema)
	}

	docSchema, err := parseDocSchema(docSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidDatabaseSchema, err)
	}

	return &CollectionSchema{
		Name:      name,
		DocSchema: docSchema,
	}, nil
}

func parseDocSchema(data map[string]any) (*DocSchema, error) {
	schemaType, ok := data["type"].(string)
	if !ok || schemaType != DOC_SCHEMA {
		return nil, fmt.Errorf("%w: invalid doc schema type", ErrInvalidDatabaseSchema)
	}

	fieldsData, ok := data["fields"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid fields data", ErrInvalidDatabaseSchema)
	}

	fields := make(map[string]any)
	for key, val := range fieldsData {
		fieldData, ok := val.(map[string]any)
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
		Fields: fields,
	}, nil
}

func parseFieldSchema(data map[string]any) (any, error) {
	schemaType, ok := data["type"].(string)
	if !ok {
		return nil, fmt.Errorf("%w: field type is required", ErrInvalidDatabaseSchema)
	}

	switch schemaType {
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
func parseAnySchema(data map[string]any) (*AnySchema, error) {
	nullable, _ := data["nullable"].(bool)
	return &AnySchema{
		Nullable: nullable,
	}, nil
}

func parseBooleanSchema(data map[string]any) (*BooleanSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &BooleanSchema{
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseDateSchema(data map[string]any) (*DateSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &DateSchema{
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseEnumSchema(data map[string]any) (*EnumSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)

	valuesInterface, ok := data["values"].([]any)
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
		Values:    values,
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseListSchema(data map[string]any) (*ListSchema, error) {
	nullable, _ := data["nullable"].(bool)

	itemSchemaData, ok := data["itemSchema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid item schema", ErrInvalidDatabaseSchema)
	}

	itemSchema, err := parseFieldSchema(itemSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing item schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &ListSchema{
		Nullable:   nullable,
		ItemSchema: itemSchema,
	}, nil
}

func parseMovableListSchema(data map[string]any) (*MovableListSchema, error) {
	nullable, _ := data["nullable"].(bool)

	itemSchemaData, ok := data["itemSchema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid item schema", ErrInvalidDatabaseSchema)
	}

	itemSchema, err := parseFieldSchema(itemSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing item schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &MovableListSchema{
		Nullable:   nullable,
		ItemSchema: itemSchema,
	}, nil
}

func parseNumberSchema(data map[string]any) (*NumberSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &NumberSchema{
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseObjectSchema(data map[string]any) (*ObjectSchema, error) {
	nullable, _ := data["nullable"].(bool)

	shapeData, ok := data["shape"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid shape data", ErrInvalidDatabaseSchema)
	}

	shape := make(map[string]any)
	for key, val := range shapeData {
		fieldData, ok := val.(map[string]any)
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
		Nullable: nullable,
		Shape:    shape,
	}, nil
}

func parseRecordSchema(data map[string]any) (*RecordSchema, error) {
	nullable, _ := data["nullable"].(bool)

	valueSchemaData, ok := data["valueSchema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid value schema", ErrInvalidDatabaseSchema)
	}

	valueSchema, err := parseFieldSchema(valueSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing value schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &RecordSchema{
		Nullable:    nullable,
		ValueSchema: valueSchema,
	}, nil
}

func parseStringSchema(data map[string]any) (*StringSchema, error) {
	nullable, _ := data["nullable"].(bool)
	unique, _ := data["unique"].(bool)
	indexType, _ := data["indexType"].(string)
	return &StringSchema{
		Nullable:  nullable,
		Unique:    unique,
		IndexType: IndexType(indexType),
	}, nil
}

func parseTextSchema(data map[string]any) (*TextSchema, error) {
	nullable, _ := data["nullable"].(bool)
	indexType, _ := data["indexType"].(string)
	return &TextSchema{
		Nullable:  nullable,
		IndexType: IndexType(indexType),
	}, nil
}

func parseTreeSchema(data map[string]any) (*TreeSchema, error) {
	nullable, _ := data["nullable"].(bool)

	treeNodeSchemaData, ok := data["treeNodeSchema"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: invalid tree node schema", ErrInvalidDatabaseSchema)
	}

	treeNodeSchema, err := parseFieldSchema(treeNodeSchemaData)
	if err != nil {
		return nil, fmt.Errorf("%w: error parsing tree node schema: %v", ErrInvalidDatabaseSchema, err)
	}

	return &TreeSchema{
		Nullable:       nullable,
		TreeNodeSchema: treeNodeSchema,
	}, nil
}

func (d *DatabaseSchema) ToJSON() map[string]any {
	collections := make(map[string]any)
	for name, coll := range d.Collections {
		collections[name] = coll.ToJSON()
	}

	return map[string]any{
		"type":        DATABASE_SCHEMA,
		"name":        d.Name,
		"version":     d.Version,
		"collections": collections,
	}
}

func (c *CollectionSchema) ToJSON() map[string]any {
	return map[string]any{
		"type":      COLLECTION_SCHEMA,
		"name":      c.Name,
		"docSchema": c.DocSchema.ToJSON(),
	}
}

func (d *DocSchema) ToJSON() map[string]any {
	fields := make(map[string]any)
	for name, field := range d.Fields {
		if s, ok := field.(interface{ ToJSON() map[string]any }); ok {
			fields[name] = s.ToJSON()
		}
	}

	return map[string]any{
		"type":   DOC_SCHEMA,
		"fields": fields,
	}
}

func (a *AnySchema) ToJSON() map[string]any {
	return map[string]any{
		"type":     ANY_SCHEMA,
		"nullable": a.Nullable,
	}
}

func (b *BooleanSchema) ToJSON() map[string]any {
	return map[string]any{
		"type":      BOOLEAN_SCHEMA,
		"nullable":  b.Nullable,
		"unique":    b.Unique,
		"indexType": b.IndexType,
	}
}

func (d *DateSchema) ToJSON() map[string]any {
	return map[string]any{
		"type":      DATE_SCHEMA,
		"nullable":  d.Nullable,
		"unique":    d.Unique,
		"indexType": d.IndexType,
	}
}

func (e *EnumSchema) ToJSON() map[string]any {
	return map[string]any{
		"type":      ENUM_SCHEMA,
		"values":    e.Values,
		"nullable":  e.Nullable,
		"unique":    e.Unique,
		"indexType": e.IndexType,
	}
}

func (l *ListSchema) ToJSON() map[string]any {
	item := l.ItemSchema.(interface{ ToJSON() map[string]any }).ToJSON()
	return map[string]any{
		"type":       LIST_SCHEMA,
		"nullable":   l.Nullable,
		"itemSchema": item,
	}
}

func (m *MovableListSchema) ToJSON() map[string]any {
	item := m.ItemSchema.(interface{ ToJSON() map[string]any }).ToJSON()
	return map[string]any{
		"type":       MOVABLE_LIST_SCHEMA,
		"nullable":   m.Nullable,
		"itemSchema": item,
	}
}

func (n *NumberSchema) ToJSON() map[string]any {
	return map[string]any{
		"type":      NUMBER_SCHEMA,
		"nullable":  n.Nullable,
		"unique":    n.Unique,
		"indexType": n.IndexType,
	}
}

func (o *ObjectSchema) ToJSON() map[string]any {
	shape := make(map[string]any)
	for name, field := range o.Shape {
		if s, ok := field.(interface{ ToJSON() map[string]any }); ok {
			shape[name] = s.ToJSON()
		}
	}
	return map[string]any{
		"type":     OBJECT_SCHEMA,
		"nullable": o.Nullable,
		"shape":    shape,
	}
}

func (r *RecordSchema) ToJSON() map[string]any {
	value := r.ValueSchema.(interface{ ToJSON() map[string]any }).ToJSON()
	return map[string]any{
		"type":        RECORD_SCHEMA,
		"nullable":    r.Nullable,
		"valueSchema": value,
	}
}

func (s *StringSchema) ToJSON() map[string]any {
	return map[string]any{
		"type":      STRING_SCHEMA,
		"nullable":  s.Nullable,
		"unique":    s.Unique,
		"indexType": s.IndexType,
	}
}

func (t *TextSchema) ToJSON() map[string]any {
	return map[string]any{
		"type":      TEXT_SCHEMA,
		"nullable":  t.Nullable,
		"indexType": t.IndexType,
	}
}

func (t *TreeSchema) ToJSON() map[string]any {
	node := t.TreeNodeSchema.(interface{ ToJSON() map[string]any }).ToJSON()
	return map[string]any{
		"type":           TREE_SCHEMA,
		"nullable":       t.Nullable,
		"treeNodeSchema": node,
	}
}
