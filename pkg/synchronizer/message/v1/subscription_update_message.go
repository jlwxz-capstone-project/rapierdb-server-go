package message

import (
	"bytes"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/query"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/synchronizer/subscription"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type SubscriptionUpdateMessageV1 struct {
	Added   []subscription.Subscription
	Removed []subscription.Subscription
}

func (m *SubscriptionUpdateMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, uint64(len(m.Added)))
	for _, sub := range m.Added {
		util.WriteVarString(buf, sub.Collection)
		queryBytes, err := sub.Query.MarshalJSON()
		if err != nil {
			return nil, err
		}
		util.WriteBytes(buf, queryBytes)
	}
	util.WriteVarUint(buf, uint64(len(m.Removed)))
	for _, sub := range m.Removed {
		util.WriteVarString(buf, sub.Collection)
		queryBytes, err := sub.Query.MarshalJSON()
		if err != nil {
			return nil, err
		}
		util.WriteBytes(buf, queryBytes)
	}
	return buf.Bytes(), nil
}

func DecodeSubscriptionUpdateMessageV1(b *bytes.Buffer) (*SubscriptionUpdateMessageV1, error) {
	addedLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	added := make([]subscription.Subscription, 0, addedLen)
	for i := uint64(0); i < addedLen; i++ {
		collection, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		queryBytes, err := util.ReadBytes(b)
		if err != nil {
			return nil, err
		}
		var query query.Query
		err = query.UnmarshalJSON(queryBytes)
		if err != nil {
			return nil, err
		}
		added = append(added, subscription.Subscription{
			Collection: collection,
			Query:      query,
		})
	}

	removedLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	removed := make([]subscription.Subscription, 0, removedLen)
	for i := uint64(0); i < removedLen; i++ {
		collection, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		queryBytes, err := util.ReadBytes(b)
		if err != nil {
			return nil, err
		}
		var query query.Query
		err = query.UnmarshalJSON(queryBytes)
		if err != nil {
			return nil, err
		}
		removed = append(removed, subscription.Subscription{
			Collection: collection,
			Query:      query,
		})
	}

	return &SubscriptionUpdateMessageV1{
		Added:   added,
		Removed: removed,
	}, nil
}
