package message

type TransactionFailedMessage struct {
	TxID   string
	Reason error
}
