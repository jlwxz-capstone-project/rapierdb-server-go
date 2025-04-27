package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// VersionQueryRespMessageV1 由客户端发送给服务端
// 用于响应服务端的 VersionQueryMessage，携带有服务端要求的
// 指定文档的版本
type VersionQueryRespMessageV1 struct {
	Responses map[string][]byte // docKey -> version (bytes)，空数组表示文档不存在
}

var _ Message = &VersionQueryRespMessageV1{}

func (m *VersionQueryRespMessageV1) isMessage() {}

func (m *VersionQueryRespMessageV1) DebugPrint() string {
	respStrs := make([]string, len(m.Responses))
	i := 0
	for docKey, version := range m.Responses {
		docKeyBytes := util.String2Bytes(docKey)
		collection := storage_engine.GetCollectionNameFromKey(docKeyBytes)
		docId := storage_engine.GetDocIdFromKey(docKeyBytes)
		respStrs[i] = fmt.Sprintf("[%s.%s]: %s", collection, docId, version)
		i++
	}
	return fmt.Sprintf("VersionQueryRespMessageV1{Responses: {%s}}", strings.Join(respStrs, ", "))
}

// decodeVersionQueryRespMessageV1Body 从 bytes.Buffer 中解码得到 VersionQueryRespMessageV1
// 如果解码失败，返回 nil
func decodeVersionQueryRespMessageV1Body(b *bytes.Buffer) (*VersionQueryRespMessageV1, error) {
	nDocs, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}
	responses := make(map[string][]byte)
	for i := uint64(0); i < nDocs; i++ {
		docKey, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		version, err := util.ReadVarByteArray(b)
		if err != nil {
			return nil, err
		}
		responses[docKey] = version
	}
	return &VersionQueryRespMessageV1{
		Responses: responses,
	}, nil
}

// Encode 将 VersionQueryRespMessageV1 编码为 []byte
func (m *VersionQueryRespMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	util.WriteVarUint(buf, uint64(len(m.Responses)))
	for docKey, version := range m.Responses {
		util.WriteVarString(buf, docKey)
		util.WriteVarByteArray(buf, version)
	}
	return buf.Bytes(), nil
}

func (m *VersionQueryRespMessageV1) Type() uint64 {
	return MSG_TYPE_VERSION_QUERY_RESP_V1
}
