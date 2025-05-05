package message

import (
	"bytes"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/db_conn"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// PostTransactionMessageV1 由客户端发送给服务端
// 表示客户端提交的事务
type PostTransactionMessageV1 struct {
	Transaction *db_conn.Transaction
}

var _ Message = &PostTransactionMessageV1{}

func (m *PostTransactionMessageV1) isMessage() {}

func (m *PostTransactionMessageV1) DebugSprint() string {
	return fmt.Sprintf("PostTransactionMessageV1{Transaction: %v}", m.Transaction)
}

func (m *PostTransactionMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteUint8(buf, uint8(m.Type()))
	msgBytes, err := db_conn.EncodeTransaction(m.Transaction)
	if err != nil {
		return nil, err
	}
	buf.Write(msgBytes)
	return buf.Bytes(), nil
}

func decodePostTransactionMessageV1(b *bytes.Buffer) (*PostTransactionMessageV1, error) {
	tr, err := db_conn.ReadTransaction(b)
	if err != nil {
		return nil, err
	}
	return &PostTransactionMessageV1{
		Transaction: tr,
	}, nil
}

func (m *PostTransactionMessageV1) Type() uint8 {
	return MSG_TYPE_POST_TRANSACTION_V1
}
