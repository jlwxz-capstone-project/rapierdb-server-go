package db_conn

import (
	"bytes"
	"encoding/json"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type DatabaseMeta struct {
	// Database schema, includes
	// - name of database
	// - version of database
	// - schema of all collections
	databaseSchema *DatabaseSchema
	// Permission definition in Js
	permissionJs string
	createdAt    uint64
}

// ToBytes serializes the database meta to bytes
func (s *DatabaseMeta) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	schemaJson := s.databaseSchema.ToJSON()
	schemaJsonStr, err := json.Marshal(schemaJson)
	if err != nil {
		return nil, err
	}
	util.WriteVarByteArray(&buf, schemaJsonStr)
	util.WriteVarString(&buf, s.permissionJs)
	util.WriteUint64(&buf, s.createdAt)
	return buf.Bytes(), nil
}

// NewDatabaseMetaFromBytes 从字节数组反序列化数据库元数据
func NewDatabaseMetaFromBytes(data []byte) (*DatabaseMeta, error) {
	buf := bytes.NewBuffer(data)
	schemaBytes, err := util.ReadVarByteArray(buf)
	if err != nil {
		return nil, err
	}
	var schemaJson map[string]any
	err = json.Unmarshal(schemaBytes, &schemaJson)
	if err != nil {
		return nil, err
	}
	schema, err := NewDatabaseSchemaFromJSON(schemaJson)
	if err != nil {
		return nil, err
	}
	permissionsJs, err := util.ReadVarString(buf)
	if err != nil {
		return nil, err
	}
	createdAt, err := util.ReadUint64(buf)
	if err != nil {
		return nil, err
	}
	return &DatabaseMeta{
		databaseSchema: schema,
		permissionJs:   permissionsJs,
		createdAt:      createdAt,
	}, nil
}

// GetCollectionNames 获取数据库中所有集合的名字
func (s *DatabaseMeta) GetCollectionNames() []string {
	collections := make([]string, 0, len(s.databaseSchema.Collections))
	for collectionName := range s.databaseSchema.Collections {
		collections = append(collections, collectionName)
	}
	return collections
}

func (s *DatabaseMeta) GetPermissionJs() string {
	return s.permissionJs
}
