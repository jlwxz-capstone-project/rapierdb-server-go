package query_filter_expr

import (
	"encoding/json"
	"fmt"

	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/loro"
	"github.com/jlwxz-capstone-project/rapierdb-server-go/pkg/util"
	pe "github.com/pkg/errors"
)

// SizeExpr 检查数组的长度
type SizeExpr struct {
	Type   QueryFilterExprType `json:"type"`
	Target QueryFilterExpr     `json:"target"`
	Size   QueryFilterExpr     `json:"size"`
}

func NewSizeExpr(target QueryFilterExpr, size QueryFilterExpr) *SizeExpr {
	return &SizeExpr{
		Type:   ExprTypeSize,
		Target: target,
		Size:   size,
	}
}

func (e *SizeExpr) DebugSprint() string {
	return fmt.Sprintf("SizeExpr{Target: %s, Size: %s}", e.Target.DebugSprint(), e.Size.DebugSprint())
}

func (e *SizeExpr) Eval(doc *loro.LoroDoc) (*ValueExpr, error) {
	// 评估字段表达式
	target, err := e.Target.Eval(doc)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "evaluating target in SIZE: %v", err)
	}

	// 检查字段是否为数组
	arr, ok := target.Value.([]any)
	if !ok {
		return nil, pe.Wrapf(ErrTypeError, "expected array in SIZE expression, got %T", target.Value)
	}

	// 评估大小表达式
	size, err := e.Size.Eval(doc)
	if err != nil {
		return nil, pe.Wrapf(ErrEvalError, "evaluating size in SIZE: %v", err)
	}

	intSize := util.ToInt(size.Value)
	return NewValueExpr(len(arr) == intSize), nil
}

func (e *SizeExpr) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func newSizeExprFromJson(msg json.RawMessage) (*SizeExpr, error) {
	var temp struct {
		Type   QueryFilterExprType `json:"type"`
		Target json.RawMessage     `json:"target"`
		Size   json.RawMessage     `json:"size"`
	}

	if err := json.Unmarshal(msg, &temp); err != nil {
		return nil, err
	}

	target, err := NewQueryFilterExprFromJson(temp.Target)
	if err != nil {
		return nil, err
	}

	size, err := NewQueryFilterExprFromJson(temp.Size)
	if err != nil {
		return nil, err
	}

	return NewSizeExpr(target, size), nil
}
