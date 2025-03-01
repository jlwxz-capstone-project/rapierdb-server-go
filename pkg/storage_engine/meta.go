package storage_engine

import (
	"bytes"
	"encoding/json"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type DatabaseMeta struct {
	DatabaseSchema *DatabaseSchema
	// 权限定义的 Js 代码
	Permissions string
	CreatedAt   uint64
}

// ToBytes 将数据库元数据序列化为字节数组
func (s *DatabaseMeta) ToBytes() ([]byte, error) {
	var buf bytes.Buffer
	schemaJson := s.DatabaseSchema.ToJSON()
	schemaJsonStr, err := json.Marshal(schemaJson)
	if err != nil {
		return nil, err
	}
	util.WriteBytes(&buf, schemaJsonStr)
	util.WriteVarString(&buf, s.Permissions)
	util.WriteUint64(&buf, s.CreatedAt)
	return buf.Bytes(), nil
}

// NewDatabaseMetaFromBytes 从字节数组反序列化数据库元数据
func NewDatabaseMetaFromBytes(data []byte) (*DatabaseMeta, error) {
	buf := bytes.NewBuffer(data)
	schemaBytes, err := util.ReadBytes(buf)
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
		DatabaseSchema: schema,
		Permissions:    permissionsJs,
		CreatedAt:      createdAt,
	}, nil
}

// GetCollectionNames 获取数据库中所有集合的名字
func (s *DatabaseMeta) GetCollectionNames() []string {
	collections := make([]string, 0, len(s.DatabaseSchema.Collections))
	for collectionName := range s.DatabaseSchema.Collections {
		collections = append(collections, collectionName)
	}
	return collections
}
