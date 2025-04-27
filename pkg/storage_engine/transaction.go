package storage_engine

import (
	"bytes"
	"errors"
	"fmt"

	pe "github.com/pkg/errors"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
)

type TransactionOpType int

const (
	OP_INSERT TransactionOpType = 0
	OP_UPDATE TransactionOpType = 1
	OP_DELETE TransactionOpType = 2
)

var (
	ErrTransactionInvalid = errors.New("transaction invalid")
)

type TransactionOp interface {
	isTransactionOp()
}

type InsertOp struct {
	Collection string
	DocID      string
	Snapshot   []byte
}

func (op *InsertOp) isTransactionOp() {}

type UpdateOp struct {
	Collection string
	DocID      string
	Update     []byte
}

func (op *UpdateOp) isTransactionOp() {}

type DeleteOp struct {
	Collection string
	DocID      string
}

func (op *DeleteOp) isTransactionOp() {}

type Transaction struct {
	// 事务 ID，应该是 UUID，通常由客户端生成
	TxID string
	// 目标数据库
	TargetDatabase string
	// 提交者，通常是客户端的 ID
	Committer string
	// 操作
	Operations []TransactionOp
}

func EnsureTransactionValid(tr *Transaction) error {
	if len(tr.TxID) != 36 {
		return pe.WithStack(fmt.Errorf("%w: invalid tx_id = \"%s\"", ErrTransactionInvalid, tr.TxID))
	}

	if tr.TargetDatabase == "" {
		return pe.WithStack(fmt.Errorf("%w: invalid target_database = \"%s\"", ErrTransactionInvalid, tr.TargetDatabase))
	}

	if tr.Committer == "" {
		return pe.WithStack(fmt.Errorf("%w: invalid committer = \"%s\"", ErrTransactionInvalid, tr.Committer))
	}

	if len(tr.Operations) == 0 {
		return pe.WithStack(fmt.Errorf("%w: empty operations", ErrTransactionInvalid))
	}

	for _, op := range tr.Operations {
		switch op := op.(type) {
		case *InsertOp, *UpdateOp, *DeleteOp:
			continue
		default:
			return pe.WithStack(fmt.Errorf("%w: invalid operation type = %T", ErrTransactionInvalid, op))
		}
	}

	return nil
}

func writeInsertOp(buf *bytes.Buffer, op *InsertOp) error {
	if err := util.WriteUint8(buf, uint8(OP_INSERT)); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, op.Collection); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, op.DocID); err != nil {
		return err
	}
	if err := util.WriteVarByteArray(buf, op.Snapshot); err != nil {
		return err
	}
	return nil
}

func writeUpdateOp(buf *bytes.Buffer, op *UpdateOp) error {
	if err := util.WriteUint8(buf, uint8(OP_UPDATE)); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, op.Collection); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, op.DocID); err != nil {
		return err
	}
	if err := util.WriteVarByteArray(buf, op.Update); err != nil {
		return err
	}
	return nil
}

func writeDeleteOp(buf *bytes.Buffer, op *DeleteOp) error {
	if err := util.WriteUint8(buf, uint8(OP_DELETE)); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, op.Collection); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, op.DocID); err != nil {
		return err
	}
	return nil
}

func writeTransactionOp(buf *bytes.Buffer, op TransactionOp) error {
	switch op := op.(type) {
	case *InsertOp:
		return writeInsertOp(buf, op)
	case *UpdateOp:
		return writeUpdateOp(buf, op)
	case *DeleteOp:
		return writeDeleteOp(buf, op)
	default:
		return pe.Errorf("unknown operation type: %T", op)
	}
}
func writeTransaction(buf *bytes.Buffer, tr *Transaction) error {
	if err := util.WriteVarString(buf, tr.TxID); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, tr.TargetDatabase); err != nil {
		return err
	}
	if err := util.WriteVarString(buf, tr.Committer); err != nil {
		return err
	}
	if err := util.WriteVarUint(buf, uint64(len(tr.Operations))); err != nil {
		return err
	}
	for _, op := range tr.Operations {
		if err := writeTransactionOp(buf, op); err != nil {
			return err
		}
	}
	return nil
}

func EncodeTransaction(tr *Transaction) (res []byte, err error) {
	buf := bytes.NewBuffer(nil)
	err = writeTransaction(buf, tr)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readTransactionOp(data *bytes.Buffer) (TransactionOp, error) {
	opType, err := util.ReadUint8(data)
	if err != nil {
		return nil, err
	}

	switch TransactionOpType(opType) {
	case OP_INSERT:
		collection, err := util.ReadVarString(data)
		if err != nil {
			return nil, err
		}
		docID, err := util.ReadVarString(data)
		if err != nil {
			return nil, err
		}
		snapshot, err := util.ReadVarByteArray(data)
		if err != nil {
			return nil, err
		}
		return &InsertOp{
			Collection: collection,
			DocID:      docID,
			Snapshot:   snapshot,
		}, nil
	case OP_UPDATE:
		collection, err := util.ReadVarString(data)
		if err != nil {
			return nil, err
		}
		docID, err := util.ReadVarString(data)
		if err != nil {
			return nil, err
		}
		update, err := util.ReadVarByteArray(data)
		if err != nil {
			return nil, err
		}
		return &UpdateOp{
			Collection: collection,
			DocID:      docID,
			Update:     update,
		}, nil
	case OP_DELETE:
		collection, err := util.ReadVarString(data)
		if err != nil {
			return nil, err
		}
		docID, err := util.ReadVarString(data)
		if err != nil {
			return nil, err
		}
		return &DeleteOp{
			Collection: collection,
			DocID:      docID,
		}, nil
	default:
		return nil, pe.Errorf("unknown operation type: %d", opType)
	}
}

func ReadTransaction(data *bytes.Buffer) (*Transaction, error) {
	txID, err := util.ReadVarString(data)
	if err != nil {
		return nil, err
	}
	targetDatabase, err := util.ReadVarString(data)
	if err != nil {
		return nil, err
	}
	committer, err := util.ReadVarString(data)
	if err != nil {
		return nil, err
	}
	nOperations, err := util.ReadVarUint(data)
	if err != nil {
		return nil, err
	}

	operations := make([]TransactionOp, 0, nOperations)
	for i := uint64(0); i < nOperations; i++ {
		op, err := readTransactionOp(data)
		if err != nil {
			return nil, err // Propagate error
		}
		operations = append(operations, op)
	}

	return &Transaction{
		TxID:           txID,
		TargetDatabase: targetDatabase,
		Committer:      committer,
		Operations:     operations,
	}, nil
}

func DecodeTransaction(data []byte) (res *Transaction, err error) {
	buf := bytes.NewBuffer(data)
	res, err = ReadTransaction(buf)
	return res, err
}
