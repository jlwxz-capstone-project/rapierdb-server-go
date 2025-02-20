package message

import (
	"bytes"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// SyncMessage 由服务端发送给客户端
// 携带有客户端需要同步的数据
type SyncMessageV1 struct {
	Upsert map[string][]byte
	Delete []string
}

func (m *SyncMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteVarUint(buf, uint64(len(m.Upsert)))
	if err != nil {
		return nil, err
	}
	for key, value := range m.Upsert {
		err = util.WriteVarString(buf, key)
		if err != nil {
			return nil, err
		}
		err = util.WriteBytes(buf, value)
		if err != nil {
			return nil, err
		}
	}
	err = util.WriteVarUint(buf, uint64(len(m.Delete)))
	if err != nil {
		return nil, err
	}
	for _, key := range m.Delete {
		err = util.WriteVarString(buf, key)
		if err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

func DecodeSyncMessageV1(b *bytes.Buffer) (*SyncMessageV1, error) {
	upsertLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	upsert := make(map[string][]byte, upsertLen)
	for i := uint64(0); i < upsertLen; i++ {
		key, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		value, err := util.ReadBytes(b)
		if err != nil {
			return nil, err
		}
		upsert[key] = value
	}

	deleteLen, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	delete := make([]string, 0, deleteLen)
	for i := uint64(0); i < deleteLen; i++ {
		key, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		delete = append(delete, key)
	}

	return &SyncMessageV1{
		Upsert: upsert,
		Delete: delete,
	}, nil
}
