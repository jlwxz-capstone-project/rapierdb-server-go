package message

import (
	"bytes"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type SubscriptionUpdateMessageV1 struct {
	Added   []string
	Removed []string
}

var _ Message = &SubscriptionUpdateMessageV1{}

func (m *SubscriptionUpdateMessageV1) isMessage() {}

func (m *SubscriptionUpdateMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	util.WriteVarUint(buf, uint64(len(m.Added)))
	for _, sub := range m.Added {
		util.WriteVarString(buf, sub)
	}
	util.WriteVarUint(buf, uint64(len(m.Removed)))
	for _, sub := range m.Removed {
		util.WriteVarString(buf, sub)
	}
	return buf.Bytes(), nil
}

func decodeSubscriptionUpdateMessageV1Body(b *bytes.Buffer) (*SubscriptionUpdateMessageV1, error) {
	addedLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	added := make([]string, 0, addedLen)
	for i := uint64(0); i < addedLen; i++ {
		collection, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		added = append(added, collection)
	}

	removedLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	removed := make([]string, 0, removedLen)
	for i := uint64(0); i < removedLen; i++ {
		collection, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		removed = append(removed, collection)
	}

	return &SubscriptionUpdateMessageV1{
		Added:   added,
		Removed: removed,
	}, nil
}

func (m *SubscriptionUpdateMessageV1) Type() uint64 {
	return MSG_TYPE_SUBSCRIPTION_UPDATE_V1
}
