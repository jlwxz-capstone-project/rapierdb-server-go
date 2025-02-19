package message

type AckTransactionMessage struct {
	TxID string
}

// DecodeAckTransactionMessage 从 []byte 中解码得到 DecodeTransactionMessage
// 如果解码失败，返回 nil
func DecodeAckTransactionMessage(bytes []byte) *AckTransactionMessage {
	return nil
}
