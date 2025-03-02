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

type InsertOp struct {
	Collection string
	DocID      string
	Snapshot   []byte
}

type UpdateOp struct {
	Collection string
	DocID      string
	Update     []byte
}

type DeleteOp struct {
	Collection string
	DocID      string
}

type Transaction struct {
	// 事务 ID，应该是 UUID，通常由客户端生成
	TxID string
	// 目标数据库
	TargetDatabase string
	// 提交者，通常是客户端的 ID
	Committer string
	// 操作
	Operations []any
}

type TransactionOp interface {
	InsertOp | UpdateOp | DeleteOp
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
		case InsertOp, UpdateOp, DeleteOp:
			continue
		default:
			return pe.WithStack(fmt.Errorf("%w: invalid operation type = %T", ErrTransactionInvalid, op))
		}
	}

	return nil
}

// Encode 将 InsertOp 编码为字节流
func (op *InsertOp) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteVarUint(buf, uint64(OP_INSERT))
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, op.Collection)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, op.DocID)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteBytes(buf, op.Snapshot)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	return buf.Bytes(), nil
}

// Encode 将 UpdateOp 编码为字节流
func (op *UpdateOp) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteVarUint(buf, uint64(OP_UPDATE))
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, op.Collection)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, op.DocID)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteBytes(buf, op.Update)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	return buf.Bytes(), nil
}

// Encode 将 DeleteOp 编码为字节流
func (op *DeleteOp) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteVarUint(buf, uint64(OP_DELETE))
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, op.Collection)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, op.DocID)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	return buf.Bytes(), nil
}

// Encode 将 Transaction 编码为字节流
func (tr *Transaction) Encode() ([]byte, error) {
	buf := &bytes.Buffer{}
	err := util.WriteVarString(buf, tr.TxID)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, tr.TargetDatabase)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarString(buf, tr.Committer)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	err = util.WriteVarUint(buf, uint64(len(tr.Operations)))
	if err != nil {
		return nil, pe.WithStack(err)
	}
	for _, op := range tr.Operations {
		var opBytes []byte
		switch op := op.(type) {
		case *InsertOp:
			opBytes, err = op.Encode()
			if err != nil {
				return nil, pe.WithStack(err)
			}
		case *UpdateOp:
			opBytes, err = op.Encode()
			if err != nil {
				return nil, pe.WithStack(err)
			}
		case *DeleteOp:
			opBytes, err = op.Encode()
			if err != nil {
				return nil, pe.WithStack(err)
			}
		}
		err = util.WriteBytes(buf, opBytes)
		if err != nil {
			return nil, pe.WithStack(err)
		}
	}
	return buf.Bytes(), nil
}

// DecodeOperation 从字节流解码操作
func DecodeOperation(data []byte) (any, error) {
	buf := bytes.NewBuffer(data)
	opType, err := util.ReadVarUint(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}

	remainingData := buf.Bytes()
	switch TransactionOpType(opType) {
	case OP_INSERT:
		return DecodeInsertOp(remainingData)
	case OP_UPDATE:
		return DecodeUpdateOp(remainingData)
	case OP_DELETE:
		return DecodeDeleteOp(remainingData)
	default:
		return nil, pe.WithStack(fmt.Errorf("%w: unknown operation type %d", ErrTransactionInvalid, opType))
	}
}

// DecodeInsertOp 从字节流解码 InsertOp
func DecodeInsertOp(data []byte) (*InsertOp, error) {
	buf := bytes.NewBuffer(data)
	collection, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	docID, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	snapshot, err := util.ReadBytes(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	return &InsertOp{
		Collection: collection,
		DocID:      docID,
		Snapshot:   snapshot,
	}, nil
}

// DecodeUpdateOp 从字节流解码 UpdateOp
func DecodeUpdateOp(data []byte) (*UpdateOp, error) {
	buf := bytes.NewBuffer(data)
	collection, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	docID, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	update, err := util.ReadBytes(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	return &UpdateOp{
		Collection: collection,
		DocID:      docID,
		Update:     update,
	}, nil
}

// DecodeDeleteOp 从字节流解码 DeleteOp
func DecodeDeleteOp(data []byte) (*DeleteOp, error) {
	buf := bytes.NewBuffer(data)
	collection, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	docID, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	return &DeleteOp{
		Collection: collection,
		DocID:      docID,
	}, nil
}

// DecodeTransaction 从字节流解码 Transaction
func DecodeTransaction(data []byte) (*Transaction, error) {
	buf := bytes.NewBuffer(data)
	txID, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	targetDatabase, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	committer, err := util.ReadVarString(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}
	opCount, err := util.ReadVarUint(buf)
	if err != nil {
		return nil, pe.WithStack(err)
	}

	tr := &Transaction{
		TxID:           txID,
		TargetDatabase: targetDatabase,
		Committer:      committer,
		Operations:     make([]any, 0, opCount),
	}

	for i := uint64(0); i < opCount; i++ {
		opBytes, err := util.ReadBytes(buf)
		if err != nil {
			return nil, pe.WithStack(err)
		}
		op, err := DecodeOperation(opBytes)
		if err != nil {
			return nil, pe.WithStack(err)
		}
		tr.Operations = append(tr.Operations, op)
	}

	return tr, nil
}
