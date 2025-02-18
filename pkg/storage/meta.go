package storage

import (
	"bytes"
	"encoding/gob"
	"time"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/schema"
)

type DatabaseMeta struct {
	Schema    schema.DatabaseSchema
	CreatedAt uint64
}

func NewEmptyDatabaseMeta() *DatabaseMeta {
	return &DatabaseMeta{
		Schema:    schema.DatabaseSchema{},
		CreatedAt: uint64(time.Now().Unix()),
	}
}

type StorageMeta struct {
	DatabaseMetas map[string]*DatabaseMeta
	CreatedAt     uint64
}

func NewEmptyStorageMeta() *StorageMeta {
	return &StorageMeta{
		DatabaseMetas: make(map[string]*DatabaseMeta),
		CreatedAt:     uint64(time.Now().Unix()),
	}
}

func (s *StorageMeta) ToBinary() ([]byte, error) {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)
	err := enc.Encode(s)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func StorageMetaFromBinary(data []byte) (*StorageMeta, error) {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	var meta StorageMeta
	if err := dec.Decode(&meta); err != nil {
		return nil, err
	}
	return &meta, nil
}

func (s *StorageMeta) GetCollectionNames(databaseName string) []string {
	dbMeta, ok := s.DatabaseMetas[databaseName]
	if !ok {
		return nil
	}

	collections := make([]string, 0, len(dbMeta.Schema.Collections))
	for collectionName := range dbMeta.Schema.Collections {
		collections = append(collections, collectionName)
	}
	return collections
}
