package message

import (
	"bytes"
	"errors"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// TransactionFailedMessageV1 由服务端发送给客户端
// 表示客户端的事务提交失败，并附上了失败的原因
type TransactionFailedMessageV1 struct {
	TxID   string
	Reason error
}

var _ Message = &TransactionFailedMessageV1{}

func (m *TransactionFailedMessageV1) isMessage() {}

func (m *TransactionFailedMessageV1) DebugSprint() string {
	return fmt.Sprintf("TransactionFailedMessageV1{TxID: %s, Reason: %v}", m.TxID, m.Reason)
}

// Encode 将 TransactionFailedMessageV1 编码为 []byte
func (m *TransactionFailedMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteUint8(buf, m.Type())
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
func decodeTransactionFailedMessageV1(b *bytes.Buffer) (*TransactionFailedMessageV1, error) {
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

func (m *TransactionFailedMessageV1) Type() uint8 {
	return MSG_TYPE_TRANSACTION_FAILED_V1
}
