package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// SubscriptionUpdateMessageV1 用于通知客户端订阅更新
// 当客户端需要更新订阅时，会发送该消息给服务器
type SubscriptionUpdateMessageV1 struct {
	Added   []query.Query
	Removed []query.Query
}

var _ Message = &SubscriptionUpdateMessageV1{}

func (m *SubscriptionUpdateMessageV1) isMessage() {}

func (m *SubscriptionUpdateMessageV1) DebugPrint() string {
	addedStrs := make([]string, len(m.Added))
	for i, sub := range m.Added {
		addedStrs[i] = sub.DebugPrint()
	}
	addedStr := strings.Join(addedStrs, ", ")
	removedStrs := make([]string, len(m.Removed))
	for i, sub := range m.Removed {
		removedStrs[i] = sub.DebugPrint()
	}
	removedStr := strings.Join(removedStrs, ", ")
	return fmt.Sprintf("SubscriptionUpdateMessageV1{Added: [%s], Removed: [%s]}", addedStr, removedStr)
}

func (m *SubscriptionUpdateMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	util.WriteVarUint(buf, uint64(len(m.Added)))
	for _, sub := range m.Added {
		encoded, err := sub.Encode()
		if err != nil {
			return nil, err
		}
		util.WriteVarByteArray(buf, encoded)
	}
	util.WriteVarUint(buf, uint64(len(m.Removed)))
	for _, sub := range m.Removed {
		encoded, err := sub.Encode()
		if err != nil {
			return nil, err
		}
		util.WriteVarByteArray(buf, encoded)
	}
	return buf.Bytes(), nil
}

func decodeSubscriptionUpdateMessageV1Body(b *bytes.Buffer) (*SubscriptionUpdateMessageV1, error) {
	addedLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	added := make([]query.Query, 0, addedLen)
	for i := uint64(0); i < addedLen; i++ {
		queryBytes, err := util.ReadVarByteArray(b)
		if err != nil {
			return nil, err
		}
		query, err := query.DecodeQuery(queryBytes)
		if err != nil {
			return nil, err
		}
		added = append(added, query)
	}

	removedLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	removed := make([]query.Query, 0, removedLen)
	for i := uint64(0); i < removedLen; i++ {
		queryBytes, err := util.ReadVarByteArray(b)
		if err != nil {
			return nil, err
		}
		query, err := query.DecodeQuery(queryBytes)
		if err != nil {
			return nil, err
		}
		removed = append(removed, query)
	}

	return &SubscriptionUpdateMessageV1{
		Added:   added,
		Removed: removed,
	}, nil
}

func (m *SubscriptionUpdateMessageV1) Type() uint64 {
	return MSG_TYPE_SUBSCRIPTION_UPDATE_V1
}
