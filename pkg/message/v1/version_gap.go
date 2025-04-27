package message

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type VersionGapMessageV1 struct {
	Responses map[string][]byte
}

var _ Message = &VersionGapMessageV1{}

func (m *VersionGapMessageV1) isMessage() {}

func (m *VersionGapMessageV1) DebugSprint() string {
	responseStrs := make([]string, len(m.Responses))
	i := 0
	for key := range m.Responses {
		responseStrs[i] = key
		i++
	}
	responseStr := strings.Join(responseStrs, ", ")
	return fmt.Sprintf("VersionGapMessageV1{Responses: [%s]}", responseStr)
}

func (m *VersionGapMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteUint8(buf, uint8(m.Type()))
	util.WriteVarUint(buf, uint64(len(m.Responses)))
	for key, value := range m.Responses {
		util.WriteVarString(buf, key)
		util.WriteVarByteArray(buf, value)
	}
	return buf.Bytes(), nil
}

func decodeVersionGapMessageV1(b *bytes.Buffer) (*VersionGapMessageV1, error) {
	nDocs, err := util.ReadVarUint(b)
	if err != nil {
		return nil, err
	}
	responses := make(map[string][]byte)
	for i := uint64(0); i < nDocs; i++ {
		key, err := util.ReadVarString(b)
		if err != nil {
			return nil, err
		}
		value, err := util.ReadVarByteArray(b)
		if err != nil {
			return nil, err
		}
		responses[key] = value
	}
	return &VersionGapMessageV1{Responses: responses}, nil
}

func (m *VersionGapMessageV1) Type() uint8 {
	return MSG_TYPE_VERSION_GAP_V1
}
