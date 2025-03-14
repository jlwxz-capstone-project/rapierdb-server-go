package message

import (
	"bytes"
	"fmt"
	"strings"

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

func (m *PostDocMessageV1) DebugPrint() string {
	upsertStrs := make([]string, len(m.Upsert))
	i := 0
	for key := range m.Upsert {
		upsertStrs[i] = key
		i++
	}
	upsertStr := strings.Join(upsertStrs, ", ")
	deleteStr := strings.Join(m.Delete, ", ")
	return fmt.Sprintf("PostDocMessageV1{Upsert: [%s], Delete: [%s]}", upsertStr, deleteStr)
}

func (m *PostDocMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
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

func decodeSyncMessageV1Body(b *bytes.Buffer) (*PostDocMessageV1, error) {
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

	return &PostDocMessageV1{
		Upsert: upsert,
		Delete: delete,
	}, nil
}

func (m *PostDocMessageV1) Type() uint64 {
	return MSG_TYPE_POST_DOC_V1
}
