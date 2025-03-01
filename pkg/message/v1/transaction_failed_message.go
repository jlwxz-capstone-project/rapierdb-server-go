package message

import (
	"bytes"
	"errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// TransactionFailedMessageV1 由服务端发送给客户端
// 表示客户端的事务提交失败，并附上了失败的原因
type TransactionFailedMessageV1 struct {
	TxID   string
	Reason error
}

// Encode 将 TransactionFailedMessageV1 编码为 []byte
func (m *TransactionFailedMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteVarString(buf, m.TxID)
	if err != nil {
		return nil, err
	}
	err = util.WriteVarString(buf, m.Reason.Error())
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeTransactionFailedMessageV1 从 bytes.Buffer 中解码得到 TransactionFailedMessageV1
// 如果解码失败，返回 nil
func DecodeTransactionFailedMessageV1(b *bytes.Buffer) (*TransactionFailedMessageV1, error) {
	txID, err := util.ReadVarString(b)
	if err != nil {
		return nil, err
	}
	reason, err := util.ReadVarString(b)
	if err != nil {
		return nil, err
	}
	return &TransactionFailedMessageV1{
		TxID:   txID,
		Reason: errors.New(reason),
	}, nil
}
