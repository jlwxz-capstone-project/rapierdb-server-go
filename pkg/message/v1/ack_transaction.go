package message

import (
	"bytes"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// AckTransactionMessageV1 由服务端发送给客户端
// 表示服务端已经收到了客户端的事务，并成功提交
type AckTransactionMessageV1 struct {
	TxID string
}

var _ Message = &AckTransactionMessageV1{}

func (m *AckTransactionMessageV1) isMessage() {}

func (m *AckTransactionMessageV1) DebugSprint() string {
	return fmt.Sprintf("AckTransactionMessageV1{TxID: %s}", m.TxID)
}

// Encode 将 AckTransactionMessageV1 编码为 []byte
func (m *AckTransactionMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	util.WriteUint8(buf, uint8(m.Type()))
	err := util.WriteVarString(buf, m.TxID)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DecodeAckTransactionMessageV1 从 bytes.Buffer 中解码得到 DecodeTransactionMessage
// 如果解码失败，返回 nil
func decodeAckTransactionMessageV1(b *bytes.Buffer) (*AckTransactionMessageV1, error) {
	txID, err := util.ReadVarString(b)
	if err != nil {
		return nil, err
	}
	return &AckTransactionMessageV1{TxID: txID}, nil
}

func (m *AckTransactionMessageV1) Type() uint8 {
	return MSG_TYPE_ACK_TRANSACTION_V1
}
