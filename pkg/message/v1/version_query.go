package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// VersionQueryMessageV1 由服务端发送给客户端
// 表示服务端希望查询客户端指定文档的版本
type VersionQueryMessageV1 struct {
	ID      string
	Queries map[string]struct{} // collection -> doc_ids
}

var _ Message = &VersionQueryMessageV1{}

func (m *VersionQueryMessageV1) isMessage() {}

func (m *VersionQueryMessageV1) DebugPrint() string {
	queryStrs := make([]string, len(m.Queries))
	i := 0
	for docKey, version := range m.Queries {
		docKeyBytes := util.String2Bytes(docKey)
		collection := storage_engine.GetCollectionNameFromKey(docKeyBytes)
		docId := storage_engine.GetDocIdFromKey(docKeyBytes)
		queryStrs[i] = fmt.Sprintf("[%s.%s]: %s", collection, docId, version)
		i++
	}
	return fmt.Sprintf("VersionQueryMessageV1{Queries: {%s}}", strings.Join(queryStrs, ", "))
}

func NewVersionQueryMessageV1(id string) *VersionQueryMessageV1 {
	return &VersionQueryMessageV1{
		ID:      id,
		Queries: make(map[string]struct{}),
	}
}

// decodeVersionQueryMessageV1Body 从 bytes.Buffer 中解码得到 VersionQueryMessageV1
// 如果解码失败，返回 nil
func decodeVersionQueryMessageV1Body(b *bytes.Buffer) (*VersionQueryMessageV1, error) {
	id, err := util.ReadVarString(b)
	if err != nil {
		return nil, err
	}
	nDocs, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}
	queries := make(map[string]struct{})
	for i := uint64(0); i < nDocs; i++ {
		docKey, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		queries[docKey] = struct{}{}
	}
	return &VersionQueryMessageV1{
		ID:      id,
		Queries: queries,
	}, nil
}

// Encode 将 VersionQueryMessageV1 编码为 []byte
func (m *VersionQueryMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	util.WriteVarString(buf, m.ID)
	nDocs := len(m.Queries)
	util.WriteVarUint(buf, uint64(nDocs))
	for docKey := range m.Queries {
		util.WriteVarString(buf, docKey)
	}
	return buf.Bytes(), nil
}

func (m *VersionQueryMessageV1) Type() uint64 {
	return MSG_TYPE_VERSION_QUERY_V1
}

func (m *VersionQueryMessageV1) AddDoc(collection string, docId string) {
	docKeyBytes, err := storage_engine.CalcDocKey(collection, docId)
	if err != nil {
		log.Errorf("VersionQueryMessageV1: 计算文档键失败: %v", err)
		return
	}
	docKey := util.Bytes2String(docKeyBytes)
	m.Queries[docKey] = struct{}{}
}

func (m *VersionQueryMessageV1) RemoveDoc(collection string, docId string) {
	docKeyBytes, err := storage_engine.CalcDocKey(collection, docId)
	if err != nil {
		log.Errorf("VersionQueryMessageV1: 计算文档键失败: %v", err)
		return
	}
	docKey := util.Bytes2String(docKeyBytes)
	delete(m.Queries, docKey)
}

func (m *VersionQueryMessageV1) GetAllCollections() []string {
	collections := make([]string, 0, len(m.Queries))
	for docKey := range m.Queries {
		docKeyBytes := util.String2Bytes(docKey)
		collections = append(collections, storage_engine.GetCollectionNameFromKey(docKeyBytes))
	}
	return collections
}

func (m *VersionQueryMessageV1) ContainsDoc(collection string, docId string) bool {
	docKeyBytes, err := storage_engine.CalcDocKey(collection, docId)
	if err != nil {
		log.Errorf("VersionQueryMessageV1: 计算文档键失败: %v", err)
		return false
	}
	docKey := util.Bytes2String(docKeyBytes)
	_, exists := m.Queries[docKey]
	return exists
}
