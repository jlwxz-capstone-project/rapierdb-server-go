package db_conn

import (
	"bytes"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

type TransactionOpType int

const (
	OP_INSERT TransactionOpType = 0
	OP_UPDATE TransactionOpType = 1
	OP_DELETE TransactionOpType = 2
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

// writeInsertOp writes an insert operation to a buffer
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

// writeUpdateOp writes an update operation to a buffer
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

// writeDeleteOp writes a delete operation to a buffer
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

// writeTransactionOp writes a transaction operation to a buffer
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

// readTransactionOp reads a transaction operation from a buffer
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
