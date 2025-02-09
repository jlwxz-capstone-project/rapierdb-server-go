package storage

type TransactionOpType int

const (
	OpInsert TransactionOpType = 0
	OpUpdate TransactionOpType = 1
	OpDelete TransactionOpType = 2
)

type InsertOp struct {
	Database   string
	Collection string
	DocID      string
	Snapshot   []byte
}

type UpdateOp struct {
	Database   string
	Collection string
	DocID      string
	Update     []byte
}

type DeleteOp struct {
	Database   string
	Collection string
	DocID      string
}

type Transaction struct {
	Operations []TransactionOp
}

type TransactionOp struct {
	Type TransactionOpType
	InsertOp
	UpdateOp
	DeleteOp
}
