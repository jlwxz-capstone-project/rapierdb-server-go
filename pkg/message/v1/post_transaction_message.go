package message

import (
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/storage_engine"
)

// PostTransactionMessageV1 由客户端发送给服务端
// 表示客户端提交的事务
type PostTransactionMessageV1 struct {
	Transaction *storage_engine.Transaction
}

func (m *PostTransactionMessageV1) Encode() ([]byte, error) {
	return m.Transaction.Encode()
}

func DecodePostTransactionMessageV1(data []byte) (*PostTransactionMessageV1, error) {
	tr, err := storage_engine.DecodeTransaction(data)
	if err != nil {
		return nil, err
	}
	return &PostTransactionMessageV1{
		Transaction: tr,
	}, nil
}
