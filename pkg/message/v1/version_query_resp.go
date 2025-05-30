package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/key_utils"
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

func (m *VersionQueryRespMessageV1) DebugSprint() string {
	fallback := "Invalid Version Query Response Message"
	respStrs := make([]string, len(m.Responses))
	i := 0
	for docKey, version := range m.Responses {
		docKeyBytes := util.String2Bytes(docKey)
		collection, err := key_utils.GetCollectionNameFromKey(docKeyBytes)
		if err != nil {
			return fallback
		}
		docId, err := key_utils.GetDocIdFromKey(docKeyBytes)
		if err != nil {
			return fallback
		}
		respStrs[i] = fmt.Sprintf("[%s.%s]: %s", collection, docId, version)
		i++
	}
	return fmt.Sprintf("VersionQueryRespMessageV1{Responses: {%s}}", strings.Join(respStrs, ", "))
}

// Encode 将 VersionQueryRespMessageV1 编码为 []byte
func (m *VersionQueryRespMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteUint8(buf, uint8(m.Type()))
	if err != nil {
		return nil, err
	}
	err = util.WriteVarUint(buf, uint64(len(m.Responses)))
	if err != nil {
		return nil, err
	}
	for docKey, version := range m.Responses {
		err := util.WriteVarString(buf, docKey)
		if err != nil {
			return nil, err
		}
		err = util.WriteVarByteArray(buf, version)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func (m *VersionQueryRespMessageV1) Type() uint8 {
	return MSG_TYPE_VERSION_QUERY_RESP_V1
}

// decodeVersionQueryRespMessageV1Body 从 bytes.Buffer 中解码得到 VersionQueryRespMessageV1
// 如果解码失败，返回 nil
func decodeVersionQueryRespMessageV1(b *bytes.Buffer) (*VersionQueryRespMessageV1, error) {
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
