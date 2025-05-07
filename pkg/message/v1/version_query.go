package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/key_utils"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/log"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// VersionQueryMessageV1 is sent from server to client
// Indicates that the server wants to query the version of specified documents from the client
type VersionQueryMessageV1 struct {
	Queries map[string]struct{} // collection -> doc_ids
}

var _ Message = &VersionQueryMessageV1{}

func (m *VersionQueryMessageV1) isMessage() {}

func (m *VersionQueryMessageV1) DebugSprint() string {
	fallback := "Invalid Version Query Message"
	queryStrs := make([]string, len(m.Queries))
	i := 0
	for docKey, version := range m.Queries {
		docKeyBytes := util.String2Bytes(docKey)
		collection, err := key_utils.GetCollectionNameFromKey(docKeyBytes)
		if err != nil {
			return fallback
		}
		docId, err := key_utils.GetDocIdFromKey(docKeyBytes)
		if err != nil {
			return fallback
		}
		queryStrs[i] = fmt.Sprintf("[%s.%s]: %s", collection, docId, version)
		i++
	}
	return fmt.Sprintf("VersionQueryMessageV1{Queries: {%s}}", strings.Join(queryStrs, ", "))
}

func NewVersionQueryMessageV1() *VersionQueryMessageV1 {
	return &VersionQueryMessageV1{
		Queries: make(map[string]struct{}),
	}
}

// Encode encodes VersionQueryMessageV1 into []byte
func (m *VersionQueryMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteUint8(buf, m.Type())
	nDocs := len(m.Queries)
	util.WriteVarUint(buf, uint64(nDocs))
	for docKey := range m.Queries {
		util.WriteVarString(buf, docKey)
	}
	return buf.Bytes(), nil
}

// decodeVersionQueryMessageV1Body decodes VersionQueryMessageV1 from bytes.Buffer
// Returns nil if decoding fails
func decodeVersionQueryMessageV1(b *bytes.Buffer) (*VersionQueryMessageV1, error) {
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
		Queries: queries,
	}, nil
}

func (m *VersionQueryMessageV1) Type() uint8 {
	return MSG_TYPE_VERSION_QUERY_V1
}

func (m *VersionQueryMessageV1) AddDoc(collection string, docId string) {
	docKeyBytes, err := key_utils.CalcDocKey(collection, docId)
	if err != nil {
		log.Errorf("VersionQueryMessageV1: Failed to calculate document key: %v", err)
		return
	}
	docKey := util.Bytes2String(docKeyBytes)
	m.Queries[docKey] = struct{}{}
}

func (m *VersionQueryMessageV1) RemoveDoc(collection string, docId string) {
	docKeyBytes, err := key_utils.CalcDocKey(collection, docId)
	if err != nil {
		log.Errorf("VersionQueryMessageV1: Failed to calculate document key: %v", err)
		return
	}
	docKey := util.Bytes2String(docKeyBytes)
	delete(m.Queries, docKey)
}

func (m *VersionQueryMessageV1) GetAllCollections() []string {
	collections := make([]string, 0, len(m.Queries))
	for docKey := range m.Queries {
		docKeyBytes := util.String2Bytes(docKey)
		collection, err := key_utils.GetCollectionNameFromKey(docKeyBytes)
		if err != nil {
			log.Errorf("VersionQueryMessageV1: Failed to get collection name: %v", err)
			continue
		}
		collections = append(collections, collection)
	}
	return collections
}

func (m *VersionQueryMessageV1) ContainsDoc(collection string, docId string) bool {
	docKeyBytes, err := key_utils.CalcDocKey(collection, docId)
	if err != nil {
		log.Errorf("VersionQueryMessageV1: Failed to calculate document key: %v", err)
		return false
	}
	docKey := util.Bytes2String(docKeyBytes)
	_, exists := m.Queries[docKey]
	return exists
}

func (m *VersionQueryMessageV1) IsEmpty() bool {
	return len(m.Queries) == 0
}
