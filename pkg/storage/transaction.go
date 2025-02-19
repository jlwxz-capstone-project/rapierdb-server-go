package storage

import (
	"errors"
	"fmt"
)

type TransactionOpType int

const (
	OpInsert TransactionOpType = 0
	OpUpdate TransactionOpType = 1
	OpDelete TransactionOpType = 2
)

var (
	ErrTransactionInvalid = errors.New("transaction invalid")
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
	// 事务 ID，应该是 UUID，通常由客户端生成
	TxID string
	// 提交者，通常是客户端的 ID
	Committer string
	// 操作
	Operations []TransactionOp
}

type TransactionOp struct {
	Type TransactionOpType
	InsertOp
	UpdateOp
	DeleteOp
}

func EnsureTransactionValid(tr *Transaction) error {
	if len(tr.TxID) != 36 {
		return fmt.Errorf("%w: invalid tx_id = %s", ErrTransactionInvalid, tr.TxID)
	}

	if tr.Committer == "" {
		return fmt.Errorf("%w: invalid committer = %s", ErrTransactionInvalid, tr.Committer)
	}

	if len(tr.Operations) == 0 {
		return fmt.Errorf("%w: empty operations", ErrTransactionInvalid)
	}

	return nil
}
