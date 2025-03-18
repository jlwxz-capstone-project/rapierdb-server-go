package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// VersionQueryRespMessageV1 由客户端发送给服务端
// 用于响应服务端的 VersionQueryMessage，携带有服务端要求的
// 指定文档的版本
type VersionQueryRespMessageV1 struct {
	Responses map[string]map[string][]byte // collection -> doc_id -> version (bytes)，空数组表示文档不存在
}

var _ Message = &VersionQueryRespMessageV1{}

func (m *VersionQueryRespMessageV1) isMessage() {}

func (m *VersionQueryRespMessageV1) DebugPrint() string {
	responsesStrs := make([]string, len(m.Responses))
	i := 0
	for collection, docs := range m.Responses {
		docsStrs := make([]string, len(docs))
		j := 0
		for docId, version := range docs {
			var versionStr string
			if len(version) == 0 {
				versionStr = "不存在"
			} else {
				versionStr = fmt.Sprintf("%v", version)
			}
			docsStrs[j] = fmt.Sprintf("%s: %s", docId, versionStr)
			j++
		}
		docsStr := strings.Join(docsStrs, ", ")
		responsesStrs[i] = fmt.Sprintf("%s: {%s}", collection, docsStr)
		i++
	}
	responsesStr := strings.Join(responsesStrs, ", ")
	return fmt.Sprintf("VersionQueryRespMessageV1{Responses: {%s}}", responsesStr)
}

// decodeVersionQueryRespMessageV1Body 从 bytes.Buffer 中解码得到 VersionQueryRespMessageV1
// 如果解码失败，返回 nil
func decodeVersionQueryRespMessageV1Body(b *bytes.Buffer) (*VersionQueryRespMessageV1, error) {
	nCollections, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}
	responses := make(map[string]map[string][]byte)
	for i := uint64(0); i < nCollections; i++ {
		collection, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		nDocs, err := util.ReadVarUint(b)
		if err != nil {
			return nil, err
		}
		docs := make(map[string][]byte)
		for j := uint64(0); j < nDocs; j++ {
			docId, err := util.ReadVarString(b)
			if err != nil {
				return nil, err
			}
			version, err := util.ReadBytes(b)
			if err != nil {
				return nil, err
			}
			docs[docId] = version
		}
		responses[collection] = docs
	}
	return &VersionQueryRespMessageV1{
		Responses: responses,
	}, nil
}

// Encode 将 VersionQueryRespMessageV1 编码为 []byte
func (m *VersionQueryRespMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	nCollections := len(m.Responses)
	util.WriteVarUint(buf, uint64(nCollections))
	for collection, docs := range m.Responses {
		util.WriteVarString(buf, collection)
		util.WriteVarUint(buf, uint64(len(docs)))
		for docId, version := range docs {
			util.WriteVarString(buf, docId)
			util.WriteBytes(buf, version) // 空数组表示文档不存在
		}
	}
	return buf.Bytes(), nil
}

func (m *VersionQueryRespMessageV1) Type() uint64 {
	return MSG_TYPE_VERSION_QUERY_RESP_V1
}
