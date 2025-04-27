package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// SubscriptionResetMessageV1 用于通知客户端重置订阅
// 当客户端需要重置订阅时，会发送该消息给服务器
type SubscriptionResetMessageV1 struct {
	Queries []query.Query
}

var _ Message = &SubscriptionResetMessageV1{}

func (m *SubscriptionResetMessageV1) isMessage() {}

func (m *SubscriptionResetMessageV1) DebugPrint() string {
	queriesStr := make([]string, len(m.Queries))
	for i, q := range m.Queries {
		queriesStr[i] = q.DebugPrint()
	}
	return fmt.Sprintf("SubscriptionResetMessageV1{Queries: [%s]}", strings.Join(queriesStr, ", "))
}

func (m *SubscriptionResetMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	util.WriteVarUint(buf, uint64(len(m.Queries)))
	for _, q := range m.Queries {
		encoded, err := q.Encode()
		if err != nil {
			return nil, err
		}
		util.WriteVarByteArray(buf, encoded)
	}
	return buf.Bytes(), nil
}

func decodeSubscriptionResetMessageV1Body(b *bytes.Buffer) (*SubscriptionResetMessageV1, error) {
	queriesLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	queries := make([]query.Query, 0, queriesLen)
	for i := uint64(0); i < queriesLen; i++ {
		queryBytes, err := util.ReadVarByteArray(b)
		if err != nil {
			return nil, err
		}
		query, err := query.DecodeQuery(queryBytes)
		if err != nil {
			return nil, err
		}
		queries = append(queries, query)
	}
	return &SubscriptionResetMessageV1{
		Queries: queries,
	}, nil
}

func (m *SubscriptionResetMessageV1) Type() uint64 {
	return MSG_TYPE_SUBSCRIPTION_RESET_V1
}
