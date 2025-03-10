package message

import (
	"bytes"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// PostTransactionMessageV1 由客户端发送给服务端
// 表示客户端提交的事务
type PostTransactionMessageV1 struct {
	Transaction *storage_engine.Transaction
}

var _ Message = &PostTransactionMessageV1{}

func (m *PostTransactionMessageV1) isMessage() {}

func (m *PostTransactionMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteVarUint(buf, m.Type())
	msgBytes, err := m.Transaction.Encode()
	if err != nil {
		return nil, err
	}
	buf.Write(msgBytes)
	return buf.Bytes(), nil
}

func decodePostTransactionMessageV1Body(b *bytes.Buffer) (*PostTransactionMessageV1, error) {
	tr, err := storage_engine.DecodeTransaction(b.Bytes())
	if err != nil {
		return nil, err
	}
	return &PostTransactionMessageV1{
		Transaction: tr,
	}, nil
}

func (m *PostTransactionMessageV1) Type() uint64 {
	return MSG_TYPE_POST_TRANSACTION_V1
}
