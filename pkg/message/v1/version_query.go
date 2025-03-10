package message

import (
	"bytes"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// VersionQueryMessageV1 由服务端发送给客户端
// 表示服务端希望查询客户端指定文档的版本
type VersionQueryMessageV1 struct {
	Queries map[string][]string // collection -> doc_ids
}

var _ Message = &VersionQueryMessageV1{}

func (m *VersionQueryMessageV1) isMessage() {}

// decodeVersionQueryMessageV1Body 从 bytes.Buffer 中解码得到 VersionQueryMessageV1
// 如果解码失败，返回 nil
func decodeVersionQueryMessageV1Body(b *bytes.Buffer) (*VersionQueryMessageV1, error) {
	nCollections, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}
	queries := make(map[string][]string)
	for i := uint64(0); i < nCollections; i++ {
		collection, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		nDocIds, err := util.ReadVarUint(b)
		if err != nil {
			return nil, err
		}
		docIds := make([]string, 0, nDocIds)
		for j := uint64(0); j < nDocIds; j++ {
			docId, err := util.ReadVarString(b)
			if err != nil {
				return nil, err
			}
			docIds = append(docIds, docId)
		}
		queries[collection] = docIds
	}
	return &VersionQueryMessageV1{
		Queries: queries,
	}, nil
}

// Encode 将 VersionQueryMessageV1 编码为 []byte
func (m *VersionQueryMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	nCollections := len(m.Queries)
	util.WriteVarUint(buf, uint64(nCollections))
	for collection, docIds := range m.Queries {
		util.WriteVarString(buf, collection)
		util.WriteVarUint(buf, uint64(len(docIds)))
		for _, docId := range docIds {
			util.WriteVarString(buf, docId)
		}
	}
	return buf.Bytes(), nil
}

func (m *VersionQueryMessageV1) Type() uint64 {
	return MSG_TYPE_VERSION_QUERY_V1
}
