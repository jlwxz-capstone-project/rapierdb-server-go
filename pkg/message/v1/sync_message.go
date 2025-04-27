package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// PostDocMessageV1 由服务端发送给客户端
// 携带有客户端需要同步的数据
type PostDocMessageV1 struct {
	Upsert map[string][]byte
	Delete []string
}

var _ Message = &PostDocMessageV1{}

func (m *PostDocMessageV1) isMessage() {}

func (m *PostDocMessageV1) DebugSprint() string {
	upsertStrs := make([]string, len(m.Upsert))
	i := 0
	for docKey := range m.Upsert {
		collection := storage_engine.GetCollectionNameFromKey([]byte(docKey))
		docId := storage_engine.GetDocIdFromKey([]byte(docKey))
		upsertStrs[i] = fmt.Sprintf("%s.%s", collection, docId)
		i++
	}
	upsertStr := strings.Join(upsertStrs, ", ")
	deleteStr := strings.Join(m.Delete, ", ")
	return fmt.Sprintf("PostDocMessageV1{Upsert: [%s], Delete: [%s]}", upsertStr, deleteStr)
}

func (m *PostDocMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteUint8(buf, uint8(m.Type()))
	err := util.WriteVarUint(buf, uint64(len(m.Upsert)))
	if err != nil {
		return nil, err
	}
	for key, value := range m.Upsert {
		err = util.WriteVarString(buf, key)
		if err != nil {
			return nil, err
		}
		err = util.WriteVarByteArray(buf, value)
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

func decodeSyncMessageV1(b *bytes.Buffer) (*PostDocMessageV1, error) {
	nUpsert, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	upsert := make(map[string][]byte, nUpsert)
	for i := uint64(0); i < nUpsert; i++ {
		key, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		value, err := util.ReadVarByteArray(b)
		if err != nil {
			return nil, err
		}
		upsert[key] = value
	}

	nDelete, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}

	delete := make([]string, 0, nDelete)
	for i := uint64(0); i < nDelete; i++ {
		key, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		delete = append(delete, key)
	}

	return &PostDocMessageV1{
		Upsert: upsert,
		Delete: delete,
	}, nil
}

func (m *PostDocMessageV1) Type() uint8 {
	return MSG_TYPE_SYNC_V1
}
