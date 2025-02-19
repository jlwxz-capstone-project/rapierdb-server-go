package message

import (
	"bytes"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

// AckTransactionMessageV1 由服务端发送给客户端
// 表示服务端已经收到了客户端的事务，并成功提交
type AckTransactionMessageV1 struct {
	TxID string
}

// DecodeAckTransactionMessageV1 从 bytes.Buffer 中解码得到 DecodeTransactionMessage
// 如果解码失败，返回 nil
func DecodeAckTransactionMessageV1(b *bytes.Buffer) (*AckTransactionMessageV1, error) {
	txID, err := util.ReadVarString(b)
	if err != nil {
		return nil, err
	}
	return &AckTransactionMessageV1{TxID: txID}, nil
}

// Encode 将 AckTransactionMessageV1 编码为 []byte
func (m *AckTransactionMessageV1) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteVarString(buf, m.TxID)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
